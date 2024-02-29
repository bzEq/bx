// Copyright (c) 2023 Kai Luo <gluokai@gmail.com>. All rights reserved.

#include <stddef.h>
#include <stdint.h>

extern "C" {

void ByteSwap(uint8_t *__restrict__ dst, const uint8_t *__restrict__ src,
              size_t len) {
  static constexpr size_t n = sizeof(uint64_t);
  auto dst64 = reinterpret_cast<uint64_t *>(dst);
  auto src64 = reinterpret_cast<const uint64_t *>(src);
  const size_t m = len / n;
  const size_t r = m * n;
  for (size_t i = 0; i < m; ++i)
    dst64[i] = __builtin_bswap64(src64[m - 1 - i]);
  for (size_t i = 0; i < len - r; ++i)
    dst[r + i] = src[len - 1 - i];
}
}
