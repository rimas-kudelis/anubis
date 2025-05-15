use anubis::{DATA_BUFFER, DATA_LENGTH, update_nonce};
use std::boxed::Box;
use std::mem::size_of;
use std::sync::{LazyLock, Mutex};

pub static RESULT_HASH: LazyLock<Mutex<[u8; 16]>> = LazyLock::new(|| Mutex::new([0; 16]));

pub static VERIFICATION_HASH: LazyLock<Box<Mutex<[u8; 16]>>> =
    LazyLock::new(|| Box::new(Mutex::new([0; 16])));

#[unsafe(no_mangle)]
pub extern "C" fn anubis_work(_difficulty: u32, initial_nonce: u32, iterand: u32) -> u32 {
    let data = &mut DATA_BUFFER.clone();
    let mut data_len = DATA_LENGTH.lock().unwrap();

    // Ensure there's enough space in the buffer for the nonce (4 bytes)
    if *data_len + size_of::<u32>() > data.len() {
        #[cfg(target_arch = "wasm32")]
        unreachable!();
        #[cfg(not(target_arch = "wasm32"))]
        panic!("Not enough space in DATA_BUFFER to write nonce");
    }

    let mut nonce = initial_nonce;

    loop {
        let nonce_bytes = nonce.to_le_bytes();
        let start = *data_len;
        let end = start + size_of::<u32>();
        data[start..end].copy_from_slice(&nonce_bytes);

        // Update the data length
        *data_len += size_of::<u32>();
        let data_slice = &data[..*data_len];

        let result = equix::solve(data_slice).unwrap();

        if result.len() == 0 {
            nonce += iterand;
            update_nonce(nonce);
            continue;
        }

        let mut challenge = RESULT_HASH.lock().unwrap();
        challenge.copy_from_slice(&result[0].to_bytes());
        return nonce;
    }
}

#[unsafe(no_mangle)]
pub extern "C" fn anubis_validate(nonce: u32, difficulty: u32) -> bool {
    true
}

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
