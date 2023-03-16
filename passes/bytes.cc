#include <stddef.h>
#include <stdint.h>

extern "C" {

void ByteSwap(uint8_t *__restrict__ dst, const uint8_t *__restrict__ src,
              size_t len) {
  static const size_t n = sizeof(uint64_t);
  uint64_t *__restrict__ u64_dst = reinterpret_cast<uint64_t *>(dst);
  const uint64_t *__restrict__ u64_src =
      reinterpret_cast<const uint64_t *>(src);
  const size_t m = len / n;
#pragma clang loop unroll(enable)
  for (size_t i = 0; i < m; ++i)
    u64_dst[i] = __builtin_bswap64(u64_src[m - 1 - i]);
#pragma clang loop unroll(enable)
  for (size_t i = m * n; i < len; ++i)
    dst[i] = src[i];
}
}
