#[cfg(target_arch = "wasm32")]
mod hostimport {
    #[link(wasm_import_module = "anubis")]
    unsafe extern "C" {
        /// The runtime expects this function to be defined. It is called whenever the Anubis check
        /// worker processes about 1024 hashes. This can be a no-op if you want.
        fn anubis_update_nonce(nonce: u32);
    }

    /// Safe wrapper to `anubis_update_nonce`.
    pub fn update_nonce(nonce: u32) {
        unsafe {
            anubis_update_nonce(nonce);
        }
    }
}

#[cfg(not(target_arch = "wasm32"))]
mod hostimport {
    pub fn update_nonce(_nonce: u32) {
        // This is intentionally blank
    }
}

pub use hostimport::update_nonce;
