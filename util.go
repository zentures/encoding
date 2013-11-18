/*
 * Copyright (c) 2013 Zhen, LLC. http://zhen.io. All rights reserved.
 * Use of this source code is governed by the Apache 2.0 license.
 *
 */

package encoding

import (
	"fmt"
	"math/big"
)

func FloorBy(value, factor int) int {
	return value - value%factor
}

func CeilBy(value, factor int) int {
	return value + factor - value%factor
}

func LeadingBitPosition(x uint32) int32 {
	//return 32 - int32(nlz1a(x))
	return int32(bitlen(big.Word(x)))
}

func DeltaMaxBits(initoffset int32, buf []int32) int32 {
	var mask int32

	for _, v := range buf {
		mask |= v - initoffset
		initoffset = v
	}

	return LeadingBitPosition(uint32(mask))
}

func MaxBits(buf []int32) int32 {
	var mask int32

	for _, v := range buf {
		mask |= v
	}

	return LeadingBitPosition(uint32(mask))
}

func PrintInt32sInBits(buf []int32) {
	fmt.Println("                           10987654321098765432109876543210")
	for i, v := range buf {
		fmt.Printf("%4d: %20d %032b\n", i, v, uint32(v))
	}
}

func Delta(in, out []int32, offset int32) {
	for i, v := range in {
		out[i] = v - offset
		offset = v
	}
}

func InverseDelta(in, out []int32, offset int32) {
	for i, v := range in {
		out[i] = v + offset
		offset = out[i]
	}
}

// https://developers.google.com/protocol-buffers/docs/encoding#types
func ZigZagDelta(in, out []int32) {
	offset := int32(0)

	for i, v := range in {
		n := v - offset
		out[i] = (n << 1) ^ (n >> 31)
		offset = v
	}
}

func InverseZigZagDelta(in, out []int32) {
	offset := int32(0)

	for i, v := range in {
		//n := int32(uint32(v) >> 1) ^ (-(v & 1))
		n := int32(uint32(v)>>1) ^ ((v << 31) >> 31)
		out[i] = n + offset
		offset = out[i]
	}
}

// Copied from http://www.hackersdelight.org/hdcodetxt/nlz.c.txt - nlz1a
func nlz1a(x uint32) uint32 {
	var n uint32 = 0
	if x <= 0 {
		return (^x >> 26) & 32
	}

	n = 1

	if (x >> 16) == 0 {
		n = n + 16
		x = x << 16
	}
	if (x >> 24) == 0 {
		n = n + 8
		x = x << 8
	}
	if (x >> 28) == 0 {
		n = n + 4
		x = x << 4
	}
	if (x >> 30) == 0 {
		n = n + 2
		x = x << 2
	}
	n = n - (x >> 31)
	return n
}

func nlz2(x uint32) uint32 {
	var y uint32
	var n uint32 = 32

	y = x >> 16
	if y != 0 {
		n = n - 16
		x = y
	}
	y = x >> 8
	if y != 0 {
		n = n - 8
		x = y
	}
	y = x >> 4
	if y != 0 {
		n = n - 4
		x = y
	}
	y = x >> 2
	if y != 0 {
		n = n - 2
		x = y
	}
	y = x >> 1
	if y != 0 {
		return n - 2
	}
	return n - x
}

/* The following are unrolled versions, but they are probably slower due to range checks */
func UnrolledDelta128(in, out []int32, offset int32) {
	out[0] = in[0] - offset
	out[1] = in[1] - in[0]
	out[2] = in[2] - in[1]
	out[3] = in[3] - in[2]
	out[4] = in[4] - in[3]
	out[5] = in[5] - in[4]
	out[6] = in[6] - in[5]
	out[7] = in[7] - in[6]
	out[8] = in[8] - in[7]
	out[9] = in[9] - in[8]
	out[10] = in[10] - in[9]
	out[11] = in[11] - in[10]
	out[12] = in[12] - in[11]
	out[13] = in[13] - in[12]
	out[14] = in[14] - in[13]
	out[15] = in[15] - in[14]
	out[16] = in[16] - in[15]
	out[17] = in[17] - in[16]
	out[18] = in[18] - in[17]
	out[19] = in[19] - in[18]
	out[20] = in[20] - in[19]
	out[21] = in[21] - in[20]
	out[22] = in[22] - in[21]
	out[23] = in[23] - in[22]
	out[24] = in[24] - in[23]
	out[25] = in[25] - in[24]
	out[26] = in[26] - in[25]
	out[27] = in[27] - in[26]
	out[28] = in[28] - in[27]
	out[29] = in[29] - in[28]
	out[30] = in[30] - in[29]
	out[31] = in[31] - in[30]
	out[32] = in[32] - in[31]
	out[33] = in[33] - in[32]
	out[34] = in[34] - in[33]
	out[35] = in[35] - in[34]
	out[36] = in[36] - in[35]
	out[37] = in[37] - in[36]
	out[38] = in[38] - in[37]
	out[39] = in[39] - in[38]
	out[40] = in[40] - in[39]
	out[41] = in[41] - in[40]
	out[42] = in[42] - in[41]
	out[43] = in[43] - in[42]
	out[44] = in[44] - in[43]
	out[45] = in[45] - in[44]
	out[46] = in[46] - in[45]
	out[47] = in[47] - in[46]
	out[48] = in[48] - in[47]
	out[49] = in[49] - in[48]
	out[50] = in[50] - in[49]
	out[51] = in[51] - in[50]
	out[52] = in[52] - in[51]
	out[53] = in[53] - in[52]
	out[54] = in[54] - in[53]
	out[55] = in[55] - in[54]
	out[56] = in[56] - in[55]
	out[57] = in[57] - in[56]
	out[58] = in[58] - in[57]
	out[59] = in[59] - in[58]
	out[60] = in[60] - in[59]
	out[61] = in[61] - in[60]
	out[62] = in[62] - in[61]
	out[63] = in[63] - in[62]
	out[64] = in[64] - in[63]
	out[65] = in[65] - in[64]
	out[66] = in[66] - in[65]
	out[67] = in[67] - in[66]
	out[68] = in[68] - in[67]
	out[69] = in[69] - in[68]
	out[70] = in[70] - in[69]
	out[71] = in[71] - in[70]
	out[72] = in[72] - in[71]
	out[73] = in[73] - in[72]
	out[74] = in[74] - in[73]
	out[75] = in[75] - in[74]
	out[76] = in[76] - in[75]
	out[77] = in[77] - in[76]
	out[78] = in[78] - in[77]
	out[79] = in[79] - in[78]
	out[80] = in[80] - in[79]
	out[81] = in[81] - in[80]
	out[82] = in[82] - in[81]
	out[83] = in[83] - in[82]
	out[84] = in[84] - in[83]
	out[85] = in[85] - in[84]
	out[86] = in[86] - in[85]
	out[87] = in[87] - in[86]
	out[88] = in[88] - in[87]
	out[89] = in[89] - in[88]
	out[90] = in[90] - in[89]
	out[91] = in[91] - in[90]
	out[92] = in[92] - in[91]
	out[93] = in[93] - in[92]
	out[94] = in[94] - in[93]
	out[95] = in[95] - in[94]
	out[96] = in[96] - in[95]
	out[97] = in[97] - in[96]
	out[98] = in[98] - in[97]
	out[99] = in[99] - in[98]
	out[100] = in[100] - in[99]
	out[101] = in[101] - in[100]
	out[102] = in[102] - in[101]
	out[103] = in[103] - in[102]
	out[104] = in[104] - in[103]
	out[105] = in[105] - in[104]
	out[106] = in[106] - in[105]
	out[107] = in[107] - in[106]
	out[108] = in[108] - in[107]
	out[109] = in[109] - in[108]
	out[110] = in[110] - in[109]
	out[111] = in[111] - in[110]
	out[112] = in[112] - in[111]
	out[113] = in[113] - in[112]
	out[114] = in[114] - in[113]
	out[115] = in[115] - in[114]
	out[116] = in[116] - in[115]
	out[117] = in[117] - in[116]
	out[118] = in[118] - in[117]
	out[119] = in[119] - in[118]
	out[120] = in[120] - in[119]
	out[121] = in[121] - in[120]
	out[122] = in[122] - in[121]
	out[123] = in[123] - in[122]
	out[124] = in[124] - in[123]
	out[125] = in[125] - in[124]
	out[126] = in[126] - in[125]
	out[127] = in[127] - in[126]
}

func UnrolledInverseDelta128(in, out []int32, offset int32) {
	out[0] = in[0] + offset
	out[1] = in[1] + out[0]
	out[2] = in[2] + out[1]
	out[3] = in[3] + out[2]
	out[4] = in[4] + out[3]
	out[5] = in[5] + out[4]
	out[6] = in[6] + out[5]
	out[7] = in[7] + out[6]
	out[8] = in[8] + out[7]
	out[9] = in[9] + out[8]
	out[10] = in[10] + out[9]
	out[11] = in[11] + out[10]
	out[12] = in[12] + out[11]
	out[13] = in[13] + out[12]
	out[14] = in[14] + out[13]
	out[15] = in[15] + out[14]
	out[16] = in[16] + out[15]
	out[17] = in[17] + out[16]
	out[18] = in[18] + out[17]
	out[19] = in[19] + out[18]
	out[20] = in[20] + out[19]
	out[21] = in[21] + out[20]
	out[22] = in[22] + out[21]
	out[23] = in[23] + out[22]
	out[24] = in[24] + out[23]
	out[25] = in[25] + out[24]
	out[26] = in[26] + out[25]
	out[27] = in[27] + out[26]
	out[28] = in[28] + out[27]
	out[29] = in[29] + out[28]
	out[30] = in[30] + out[29]
	out[31] = in[31] + out[30]
	out[32] = in[32] + out[31]
	out[33] = in[33] + out[32]
	out[34] = in[34] + out[33]
	out[35] = in[35] + out[34]
	out[36] = in[36] + out[35]
	out[37] = in[37] + out[36]
	out[38] = in[38] + out[37]
	out[39] = in[39] + out[38]
	out[40] = in[40] + out[39]
	out[41] = in[41] + out[40]
	out[42] = in[42] + out[41]
	out[43] = in[43] + out[42]
	out[44] = in[44] + out[43]
	out[45] = in[45] + out[44]
	out[46] = in[46] + out[45]
	out[47] = in[47] + out[46]
	out[48] = in[48] + out[47]
	out[49] = in[49] + out[48]
	out[50] = in[50] + out[49]
	out[51] = in[51] + out[50]
	out[52] = in[52] + out[51]
	out[53] = in[53] + out[52]
	out[54] = in[54] + out[53]
	out[55] = in[55] + out[54]
	out[56] = in[56] + out[55]
	out[57] = in[57] + out[56]
	out[58] = in[58] + out[57]
	out[59] = in[59] + out[58]
	out[60] = in[60] + out[59]
	out[61] = in[61] + out[60]
	out[62] = in[62] + out[61]
	out[63] = in[63] + out[62]
	out[64] = in[64] + out[63]
	out[65] = in[65] + out[64]
	out[66] = in[66] + out[65]
	out[67] = in[67] + out[66]
	out[68] = in[68] + out[67]
	out[69] = in[69] + out[68]
	out[70] = in[70] + out[69]
	out[71] = in[71] + out[70]
	out[72] = in[72] + out[71]
	out[73] = in[73] + out[72]
	out[74] = in[74] + out[73]
	out[75] = in[75] + out[74]
	out[76] = in[76] + out[75]
	out[77] = in[77] + out[76]
	out[78] = in[78] + out[77]
	out[79] = in[79] + out[78]
	out[80] = in[80] + out[79]
	out[81] = in[81] + out[80]
	out[82] = in[82] + out[81]
	out[83] = in[83] + out[82]
	out[84] = in[84] + out[83]
	out[85] = in[85] + out[84]
	out[86] = in[86] + out[85]
	out[87] = in[87] + out[86]
	out[88] = in[88] + out[87]
	out[89] = in[89] + out[88]
	out[90] = in[90] + out[89]
	out[91] = in[91] + out[90]
	out[92] = in[92] + out[91]
	out[93] = in[93] + out[92]
	out[94] = in[94] + out[93]
	out[95] = in[95] + out[94]
	out[96] = in[96] + out[95]
	out[97] = in[97] + out[96]
	out[98] = in[98] + out[97]
	out[99] = in[99] + out[98]
	out[100] = in[100] + out[99]
	out[101] = in[101] + out[100]
	out[102] = in[102] + out[101]
	out[103] = in[103] + out[102]
	out[104] = in[104] + out[103]
	out[105] = in[105] + out[104]
	out[106] = in[106] + out[105]
	out[107] = in[107] + out[106]
	out[108] = in[108] + out[107]
	out[109] = in[109] + out[108]
	out[110] = in[110] + out[109]
	out[111] = in[111] + out[110]
	out[112] = in[112] + out[111]
	out[113] = in[113] + out[112]
	out[114] = in[114] + out[113]
	out[115] = in[115] + out[114]
	out[116] = in[116] + out[115]
	out[117] = in[117] + out[116]
	out[118] = in[118] + out[117]
	out[119] = in[119] + out[118]
	out[120] = in[120] + out[119]
	out[121] = in[121] + out[120]
	out[122] = in[122] + out[121]
	out[123] = in[123] + out[122]
	out[124] = in[124] + out[123]
	out[125] = in[125] + out[124]
	out[126] = in[126] + out[125]
	out[127] = in[127] + out[126]
}

func UnrolledLeadingBitFrequency128(in, freqs []int32) {
	freqs[LeadingBitPosition(uint32(in[0]))]++
	freqs[LeadingBitPosition(uint32(in[1]))]++
	freqs[LeadingBitPosition(uint32(in[2]))]++
	freqs[LeadingBitPosition(uint32(in[3]))]++
	freqs[LeadingBitPosition(uint32(in[4]))]++
	freqs[LeadingBitPosition(uint32(in[5]))]++
	freqs[LeadingBitPosition(uint32(in[6]))]++
	freqs[LeadingBitPosition(uint32(in[7]))]++
	freqs[LeadingBitPosition(uint32(in[8]))]++
	freqs[LeadingBitPosition(uint32(in[9]))]++
	freqs[LeadingBitPosition(uint32(in[10]))]++
	freqs[LeadingBitPosition(uint32(in[11]))]++
	freqs[LeadingBitPosition(uint32(in[12]))]++
	freqs[LeadingBitPosition(uint32(in[13]))]++
	freqs[LeadingBitPosition(uint32(in[14]))]++
	freqs[LeadingBitPosition(uint32(in[15]))]++
	freqs[LeadingBitPosition(uint32(in[16]))]++
	freqs[LeadingBitPosition(uint32(in[17]))]++
	freqs[LeadingBitPosition(uint32(in[18]))]++
	freqs[LeadingBitPosition(uint32(in[19]))]++
	freqs[LeadingBitPosition(uint32(in[20]))]++
	freqs[LeadingBitPosition(uint32(in[21]))]++
	freqs[LeadingBitPosition(uint32(in[22]))]++
	freqs[LeadingBitPosition(uint32(in[23]))]++
	freqs[LeadingBitPosition(uint32(in[24]))]++
	freqs[LeadingBitPosition(uint32(in[25]))]++
	freqs[LeadingBitPosition(uint32(in[26]))]++
	freqs[LeadingBitPosition(uint32(in[27]))]++
	freqs[LeadingBitPosition(uint32(in[28]))]++
	freqs[LeadingBitPosition(uint32(in[29]))]++
	freqs[LeadingBitPosition(uint32(in[30]))]++
	freqs[LeadingBitPosition(uint32(in[31]))]++
	freqs[LeadingBitPosition(uint32(in[32]))]++
	freqs[LeadingBitPosition(uint32(in[33]))]++
	freqs[LeadingBitPosition(uint32(in[34]))]++
	freqs[LeadingBitPosition(uint32(in[35]))]++
	freqs[LeadingBitPosition(uint32(in[36]))]++
	freqs[LeadingBitPosition(uint32(in[37]))]++
	freqs[LeadingBitPosition(uint32(in[38]))]++
	freqs[LeadingBitPosition(uint32(in[39]))]++
	freqs[LeadingBitPosition(uint32(in[40]))]++
	freqs[LeadingBitPosition(uint32(in[41]))]++
	freqs[LeadingBitPosition(uint32(in[42]))]++
	freqs[LeadingBitPosition(uint32(in[43]))]++
	freqs[LeadingBitPosition(uint32(in[44]))]++
	freqs[LeadingBitPosition(uint32(in[45]))]++
	freqs[LeadingBitPosition(uint32(in[46]))]++
	freqs[LeadingBitPosition(uint32(in[47]))]++
	freqs[LeadingBitPosition(uint32(in[48]))]++
	freqs[LeadingBitPosition(uint32(in[49]))]++
	freqs[LeadingBitPosition(uint32(in[50]))]++
	freqs[LeadingBitPosition(uint32(in[51]))]++
	freqs[LeadingBitPosition(uint32(in[52]))]++
	freqs[LeadingBitPosition(uint32(in[53]))]++
	freqs[LeadingBitPosition(uint32(in[54]))]++
	freqs[LeadingBitPosition(uint32(in[55]))]++
	freqs[LeadingBitPosition(uint32(in[56]))]++
	freqs[LeadingBitPosition(uint32(in[57]))]++
	freqs[LeadingBitPosition(uint32(in[58]))]++
	freqs[LeadingBitPosition(uint32(in[59]))]++
	freqs[LeadingBitPosition(uint32(in[60]))]++
	freqs[LeadingBitPosition(uint32(in[61]))]++
	freqs[LeadingBitPosition(uint32(in[62]))]++
	freqs[LeadingBitPosition(uint32(in[63]))]++
	freqs[LeadingBitPosition(uint32(in[64]))]++
	freqs[LeadingBitPosition(uint32(in[65]))]++
	freqs[LeadingBitPosition(uint32(in[66]))]++
	freqs[LeadingBitPosition(uint32(in[67]))]++
	freqs[LeadingBitPosition(uint32(in[68]))]++
	freqs[LeadingBitPosition(uint32(in[69]))]++
	freqs[LeadingBitPosition(uint32(in[70]))]++
	freqs[LeadingBitPosition(uint32(in[71]))]++
	freqs[LeadingBitPosition(uint32(in[72]))]++
	freqs[LeadingBitPosition(uint32(in[73]))]++
	freqs[LeadingBitPosition(uint32(in[74]))]++
	freqs[LeadingBitPosition(uint32(in[75]))]++
	freqs[LeadingBitPosition(uint32(in[76]))]++
	freqs[LeadingBitPosition(uint32(in[77]))]++
	freqs[LeadingBitPosition(uint32(in[78]))]++
	freqs[LeadingBitPosition(uint32(in[79]))]++
	freqs[LeadingBitPosition(uint32(in[80]))]++
	freqs[LeadingBitPosition(uint32(in[81]))]++
	freqs[LeadingBitPosition(uint32(in[82]))]++
	freqs[LeadingBitPosition(uint32(in[83]))]++
	freqs[LeadingBitPosition(uint32(in[84]))]++
	freqs[LeadingBitPosition(uint32(in[85]))]++
	freqs[LeadingBitPosition(uint32(in[86]))]++
	freqs[LeadingBitPosition(uint32(in[87]))]++
	freqs[LeadingBitPosition(uint32(in[88]))]++
	freqs[LeadingBitPosition(uint32(in[89]))]++
	freqs[LeadingBitPosition(uint32(in[90]))]++
	freqs[LeadingBitPosition(uint32(in[91]))]++
	freqs[LeadingBitPosition(uint32(in[92]))]++
	freqs[LeadingBitPosition(uint32(in[93]))]++
	freqs[LeadingBitPosition(uint32(in[94]))]++
	freqs[LeadingBitPosition(uint32(in[95]))]++
	freqs[LeadingBitPosition(uint32(in[96]))]++
	freqs[LeadingBitPosition(uint32(in[97]))]++
	freqs[LeadingBitPosition(uint32(in[98]))]++
	freqs[LeadingBitPosition(uint32(in[99]))]++
	freqs[LeadingBitPosition(uint32(in[100]))]++
	freqs[LeadingBitPosition(uint32(in[101]))]++
	freqs[LeadingBitPosition(uint32(in[102]))]++
	freqs[LeadingBitPosition(uint32(in[103]))]++
	freqs[LeadingBitPosition(uint32(in[104]))]++
	freqs[LeadingBitPosition(uint32(in[105]))]++
	freqs[LeadingBitPosition(uint32(in[106]))]++
	freqs[LeadingBitPosition(uint32(in[107]))]++
	freqs[LeadingBitPosition(uint32(in[108]))]++
	freqs[LeadingBitPosition(uint32(in[109]))]++
	freqs[LeadingBitPosition(uint32(in[110]))]++
	freqs[LeadingBitPosition(uint32(in[111]))]++
	freqs[LeadingBitPosition(uint32(in[112]))]++
	freqs[LeadingBitPosition(uint32(in[113]))]++
	freqs[LeadingBitPosition(uint32(in[114]))]++
	freqs[LeadingBitPosition(uint32(in[115]))]++
	freqs[LeadingBitPosition(uint32(in[116]))]++
	freqs[LeadingBitPosition(uint32(in[117]))]++
	freqs[LeadingBitPosition(uint32(in[118]))]++
	freqs[LeadingBitPosition(uint32(in[119]))]++
	freqs[LeadingBitPosition(uint32(in[120]))]++
	freqs[LeadingBitPosition(uint32(in[121]))]++
	freqs[LeadingBitPosition(uint32(in[122]))]++
	freqs[LeadingBitPosition(uint32(in[123]))]++
	freqs[LeadingBitPosition(uint32(in[124]))]++
	freqs[LeadingBitPosition(uint32(in[125]))]++
	freqs[LeadingBitPosition(uint32(in[126]))]++
	freqs[LeadingBitPosition(uint32(in[127]))]++
}
