pub fn sha256(bytes: String) -> String {
  sha256_ffi(bytes)
}

@external(erlang, "durable_pipeline_kernel_ffi", "sha256_hex")
fn sha256_ffi(bytes: String) -> String
