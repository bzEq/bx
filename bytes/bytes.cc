#include <stddef.h>
#include <stdint.h>

extern "C" {

void ByteSwap(void *__restrict__ dst, const void *__restrict__ src, size_t len) {
  static const size_t n = sizeof(uint64_t);
  uint64_t *__restrict__ d = (uint64_t *)dst;
  const uint64_t *__restrict__ s = (uint64_t *)src;
  size_t m = len / n;
#pragma clang loop unroll(enable)
  for (size_t i = 0; i < m; ++i)
    d[i] = __builtin_bswap64(s[m-1-i]);
#pragma clang loop unroll(enable)
  for (size_t i = m * n; i < len; ++i)
    *((char *)dst + i) = *((char *)src + i);
}
}
