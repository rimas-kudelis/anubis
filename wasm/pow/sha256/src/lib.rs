use lazy_static::lazy_static;
use sha2::{Digest, Sha256};
use std::sync::Mutex;

// Statically allocated buffers at compile time.
lazy_static! {
    /// The data buffer is a bit weird in that it doesn't have an explicit length as it can
    /// and will change depending on the challenge input that was sent by the server.
    /// However, it can only fit 4096 bytes of data (one amd64 machine page). This is
    /// slightly overkill for the purposes of an Anubis check, but it's fine to assume
    /// that the browser can afford this much ram usage.
    ///
    /// Callers should fetch the base data pointer, write up to 4096 bytes, and then
    /// `set_data_length` the number of bytes they have written
    ///
    /// This is also functionally a write-only buffer, so it doesn't really matter that
    /// the length of this buffer isn't exposed.
    static ref DATA_BUFFER: Mutex<[u8; 4096]> = Mutex::new([0; 4096]);
    static ref DATA_LENGTH: Mutex<usize> = Mutex::new(0);

    /// SHA-256 hashes are 32 bytes (256 bits). These are stored in static buffers due to the
    /// fact that you cannot easily pass data from host space to WebAssembly space.
    static ref RESULT_HASH: Mutex<[u8; 32]> = Mutex::new([0; 32]);
    static ref VERIFICATION_HASH: Mutex<[u8; 32]> = Mutex::new([0; 32]);
}

#[link(wasm_import_module = "anubis")]
unsafe extern "C" {
    /// The runtime expects this function to be defined. It is called whenever the Anubis check
    /// worker processes about 1024 hashes. This can be a no-op if you want.
    fn anubis_update_nonce(nonce: u32);
}

/// Safe wrapper to `anubis_update_nonce`.
fn update_nonce(nonce: u32) {
    unsafe {
        anubis_update_nonce(nonce);
    }
}

/// Core validation function. Compare each bit in the hash by progressively masking bits until
/// some are found to not be matching.
///
/// There are probably more clever ways to do this, likely involving lookup tables or something
/// really fun like that. However in my testing this lets us get up to 200 kilohashes per second
/// on my Ryzen 7950x3D, up from about 50 kilohashes per second in JavaScript.
fn validate(hash: &[u8], difficulty: u32) -> bool {
    let mut remaining = difficulty;
    for &byte in hash {
        // If we're out of bits to check, exit. This is all good.
        if remaining == 0 {
            break;
        }

        // If there are more than 8 bits remaining, the entire byte should be a
        // zero. This fast-path compares the byte to 0 and if it matches, subtract
        // 8 bits.
        if remaining >= 8 {
            if byte != 0 {
                return false;
            }
            remaining -= 8;
        } else {
            // Otherwise mask off individual bits and check against them.
            let mask = 0xFF << (8 - remaining);
            if (byte & mask) != 0 {
                return false;
            }
            remaining = 0;
        }
    }
    true
}

/// Computes hash for given nonce.
///
/// This differs from the JavaScript implementations by constructing the hash differently. In
/// JavaScript implementations, the SHA-256 input is the result of appending the nonce as an
/// integer to the hex-formatted challenge, eg:
///
///     sha256(`${challenge}${nonce}`);
///
/// This **does work**, however I think that this can be done a bit better by operating on the
/// challenge bytes _directly_ and treating the nonce as a salt.
///
/// The nonce is also randomly encoded in either big or little endian depending on the last
/// byte of the data buffer in an effort to make it more annoying to automate with GPUs.
fn compute_hash(nonce: u32) -> [u8; 32] {
    let data = DATA_BUFFER.lock().unwrap();
    let data_len = *DATA_LENGTH.lock().unwrap();
    let use_le = data[data_len - 1] >= 128;

    let data_slice = &data[..data_len];

    let mut hasher = Sha256::new();
    hasher.update(data_slice);
    hasher.update(if use_le {
        nonce.to_le_bytes()
    } else {
        nonce.to_be_bytes()
    });
    hasher.finalize().into()
}

/// This function is the main entrypoint for the Anubis proof of work implementation.
///
/// This expects `DATA_BUFFER` to be pre-populated with the challenge value as "raw bytes".
/// The definition of what goes in the data buffer is an exercise for the implementor, but
/// for SHA-256 we store the hash as "raw bytes". The data buffer is intentionally oversized
/// so that the challenge value can be expanded in the future.
///
/// `difficulty` is the number of leading bits that must match `0` in order for the
/// challenge to be successfully passed. This will be validated by the server.
///
/// `initial_nonce` is the initial value of the nonce (number used once). This nonce will be
/// appended to the challenge value in order to find a hash matching the specified
/// difficulty.
///
/// `iterand` (noun form of iterate) is the amount that the nonce should be increased by
/// every iteration of the proof of work loop. This will vary by how many threads are
/// running the proof-of-work check, and also functions as a thread ID. This prevents
/// wasting CPU time retrying a hash+nonce pair that likely won't work.
#[unsafe(no_mangle)]
pub extern "C" fn anubis_work(difficulty: u32, initial_nonce: u32, iterand: u32) -> u32 {
    let mut nonce = initial_nonce;

    loop {
        let hash = compute_hash(nonce);

        if validate(&hash, difficulty) {
            // If the challenge worked, copy the bytes into `RESULT_HASH` so the runtime
            // can pick it up.
            let mut challenge = RESULT_HASH.lock().unwrap();
            challenge.copy_from_slice(&hash);
            return nonce;
        }

        let old_nonce = nonce;
        nonce = nonce.wrapping_add(iterand);

        // send a progress update every 1024 iterations. since each thread checks
        // separate values, one simple way to do this is by bit masking the
        // nonce for multiples of 1024. unfortunately, if the number of threads
        // is not prime, only some of the threads will be sending the status
        // update and they will get behind the others. this is slightly more
        // complicated but ensures an even distribution between threads.
        if nonce > old_nonce | 1023 && (nonce >> 10) % iterand == initial_nonce {
            update_nonce(nonce);
        }
    }
}

/// This function is called by the server in order to validate a proof-of-work challenge.
/// This expects `DATA_BUFFER` to be set to the challenge value and `VERIFICATION_HASH` to
/// be set to the "raw bytes" of the SHA-256 hash that the client calculated.
///
/// If everything is good, it returns true. Otherwise, it returns false.
///
/// XXX(Xe): this could probably return an error code for what step fails, but this is fine
/// for now.
#[unsafe(no_mangle)]
pub extern "C" fn anubis_validate(nonce: u32, difficulty: u32) -> bool {
    let computed = compute_hash(nonce);
    let valid = validate(&computed, difficulty);
    if !valid {
        return false;
    }

    let verification = VERIFICATION_HASH.lock().unwrap();
    computed == *verification
}

// These functions exist to give pointers and lengths to the runtime around the Anubis
// checks, this allows JavaScript and Go to safely manipulate the memory layout that Rust
// has statically allocated at compile time without having to assume how the Rust compiler
// is going to lay it out.

#[unsafe(no_mangle)]
pub extern "C" fn result_hash_ptr() -> *const u8 {
    let challenge = RESULT_HASH.lock().unwrap();
    challenge.as_ptr()
}

#[unsafe(no_mangle)]
pub extern "C" fn result_hash_size() -> usize {
    RESULT_HASH.lock().unwrap().len()
}

#[unsafe(no_mangle)]
pub extern "C" fn verification_hash_ptr() -> *const u8 {
    let verification = VERIFICATION_HASH.lock().unwrap();
    verification.as_ptr()
}

#[unsafe(no_mangle)]
pub extern "C" fn verification_hash_size() -> usize {
    VERIFICATION_HASH.lock().unwrap().len()
}

#[unsafe(no_mangle)]
pub extern "C" fn data_ptr() -> *const u8 {
    let challenge = DATA_BUFFER.lock().unwrap();
    challenge.as_ptr()
}

#[unsafe(no_mangle)]
pub extern "C" fn set_data_length(len: u32) {
    let mut data_length = DATA_LENGTH.lock().unwrap();
    *data_length = len as usize;
}
