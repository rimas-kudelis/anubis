use lazy_static::lazy_static;
use sha2::{Digest, Sha256};
use std::sync::Mutex;

lazy_static! {
    static ref DATA_BUFFER: Mutex<[u8; 1024]> = Mutex::new([0; 1024]);
    static ref DATA_LENGTH: Mutex<usize> = Mutex::new(0);
    static ref RESULT_HASH: Mutex<[u8; 32]> = Mutex::new([0; 32]);
    static ref VERIFICATION_HASH: Mutex<[u8; 32]> = Mutex::new([0; 32]);
}

#[link(wasm_import_module = "anubis")] // Usually matches your JS namespace
unsafe extern "C" {
    // Declare the imported function
    fn anubis_update_nonce(nonce: u32);
}

fn update_nonce(nonce: u32) {
    unsafe {
        anubis_update_nonce(nonce);
    }
}

/// Core validation function
fn validate(hash: &[u8], difficulty: u32) -> bool {
    let mut remaining = difficulty;
    for &byte in hash {
        if remaining == 0 {
            break;
        }
        if remaining >= 8 {
            if byte != 0 {
                return false;
            }
            remaining -= 8;
        } else {
            let mask = 0xFF << (8 - remaining);
            if (byte & mask) != 0 {
                return false;
            }
            remaining = 0;
        }
    }
    true
}

/// Computes hash for given nonce
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

// WebAssembly exports

#[unsafe(no_mangle)]
pub extern "C" fn anubis_work(difficulty: u32, initial_nonce: u32, iterand: u32) -> u32 {
    let mut nonce = initial_nonce;

    loop {
        let hash = compute_hash(nonce);

        if validate(&hash, difficulty) {
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

#[unsafe(no_mangle)]
pub extern "C" fn anubis_validate(nonce: u32, difficulty: u32) -> bool {
    let computed = compute_hash(nonce);
    let valid = validate(&computed, difficulty);

    let verification = VERIFICATION_HASH.lock().unwrap();
    valid && computed == *verification
}

// Memory accessors

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
    // Add missing length setter
    let mut data_length = DATA_LENGTH.lock().unwrap();
    *data_length = len as usize;
}
