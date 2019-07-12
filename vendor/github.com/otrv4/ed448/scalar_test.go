package ed448

import . "gopkg.in/check.v1"

func (s *Ed448Suite) Test_MontgomeryMultiplication(c *C) {
	a := &scalar{
		0xd013f18b, 0xa03bc31f, 0xa5586c00, 0x5269ccea,
		0x80becb3f, 0x38058556, 0x736c3c5b, 0x07909887,
		0x87190ede, 0x2aae8688, 0x2c3dc273, 0x47cf8cac,
		0x3b089f07, 0x1e63e807,
	}
	b := &scalar{0x01}

	exp := &scalar{
		0xf19fb32f, 0x62bc6ae6, 0xed626086, 0x0e2d81d7,
		0x7a83d54b, 0x38e73799, 0x485ad3d6, 0x45399c9e,
		0x824b12d9, 0x5ae842c9, 0x5ca5b606, 0x3c0978b3,
		0x893b4262, 0x22c93812,
	}

	out := new(scalar)
	out.montgomeryMultiply(a, b)

	c.Assert(out, DeepEquals, exp)

	out.montgomeryMultiply(out, scalarR2)

	// by identity
	c.Assert(out, DeepEquals, a)

	a = &scalar{
		0xd013f18b, 0xa03bc31f, 0xa5586c00, 0x5269ccea,
		0x80becb3f, 0x38058556, 0x736c3c5b, 0x07909887,
		0x87190ede, 0x2aae8688, 0x2c3dc273, 0x47cf8cac,
		0x3b089f07, 0x1e63e807,
	}

	out = new(scalar)
	out.montgomeryMultiply(a, scalarZero)

	//by zero
	c.Assert(out, DeepEquals, scalarZero)

	x := &scalar{
		0xffb823a3, 0xc96a3c35, 0x7f8ed27d, 0x087b8fb9,
		0x1d9ac30a, 0x74d65764, 0xc0be082e, 0xa8cb0ae8,
		0xa8fa552b, 0x2aae8688, 0x2c3dc273, 0x47cf8cac,
		0x3b089f07, 0x1e63e807,
	}
	y := &scalar{
		0xd8bedc42, 0x686eb329, 0xe416b899, 0x17aa6d9b,
		0x1e30b38b, 0x188c6b1a, 0xd099595b, 0xbc343bcb,
		0x1adaa0e7, 0x24e8d499, 0x8e59b308, 0x0a92de2d,
		0xcae1cb68, 0x16c5450a,
	}

	exp = &scalar{
		0x14aec10b, 0x426d3399, 0x3f79af9e, 0xb1f67159,
		0x6aa5e214, 0x33819c2b, 0x19c30a89, 0x480bdc8b,
		0x7b3e1c0f, 0x5e01dfc8, 0x9414037f, 0x345954ce,
		0x611e7191, 0x19381160,
	}

	out = new(scalar)
	out.montgomeryMultiply(x, y)

	c.Assert(out, DeepEquals, exp)
}

func (s *Ed448Suite) Test_MontgomerySquare(c *C) {
	a := &scalar{
		0xcf5fac3d, 0x7e56a34b, 0xf640922b, 0x3fa50692,
		0x1370f8b8, 0x6f08f331, 0x8dccc486, 0x4bb395e0,
		0xf22c6951, 0x21cc3078, 0xd2391f9d, 0x930392e5,
		0x04b3273b, 0x31620816,
	}

	exp := &scalar{
		0x15598f62, 0xb9b1ed71, 0x52fcd042, 0x862a9f10,
		0x1e8a309f, 0x9988f8e0, 0xa22347d7, 0xe9ab2c22,
		0x38363f74, 0xfd7c58aa, 0xc49a1433, 0xd9a6c4c3,
		0x75d3395e, 0x0d79f6e3,
	}

	out := new(scalar)
	out.montgomerySquare(a)

	c.Assert(out, DeepEquals, exp)

}

func (s *Ed448Suite) Test_ScalarInverse(c *C) {
	a := &scalar{
		0x3ac84414, 0x2381c577, 0x765665e6, 0x7ad87d3c,
		0x1be79ea1, 0x10b1fe80, 0x73bf5fc4, 0xe892b6ec,
		0x1946e0b9, 0x97b2cb1a, 0x40f3f31a, 0xdcb2e06c,
		0x628dad63, 0x127a1c5f,
	}

	exp := &scalar{
		0x034653d2, 0x66a1783c, 0xa5ec956b, 0x30a35363,
		0x31c4586f, 0x9199bcc0, 0xeb2f34da, 0x83f624f7,
		0x8bb70775, 0x8e34702d, 0x8bcd73fd, 0xbee7c614,
		0x45874923, 0x0892726c,
	}

	ok := a.invert()

	c.Assert(a, DeepEquals, exp)
	c.Assert(ok, Equals, true)
}

func (s *Ed448Suite) Test_ScalarEquality(c *C) {
	a := &scalar{
		0xffb823a3, 0xc96a3c35, 0x7f8ed27d, 0x087b8fb9,
		0x1d9ac30a, 0x74d65764, 0xc0be082e, 0xa8cb0ae8,
		0xa8fa552b, 0x2aae8688, 0x2c3dc273, 0x47cf8cac,
		0x3b089f07, 0x1e63e807,
	}

	b := &scalar{
		0xfffd23f3, 0xcf6a4c35, 0x7f8ed27d, 0x087b8fb9,
		0x7d9ac31a, 0x74d65764, 0x30be082e, 0x68cb14e8,
		0xaaaa552b, 0x3aae8588, 0x2c3dc273, 0x68cf88ac,
		0x3b089f07, 0x1e6eee07,
	}
	c.Assert(a.equals(a), Equals, true)
	c.Assert(a.equals(b), Equals, false)
}

func (s *Ed448Suite) Test_ScalarCopy(c *C) {
	exp := &scalar{
		0xffb823a3, 0xc96a3c35, 0x7f8ed27d, 0x087b8fb9,
		0x1d9ac30a, 0x74d65764, 0xc0be082e, 0xa8cb0ae8,
		0xa8fa552b, 0x2aae8688, 0x2c3dc273, 0x47cf8cac,
		0x3b089f07, 0x1e63e807,
	}
	a := exp.copy()
	c.Assert(a, DeepEquals, exp)
}

func (s *Ed448Suite) Test_ScalarSet(c *C) {
	a := &scalar{}
	a.set(word(0xee))

	exp := &scalar{0xee}

	c.Assert(a, DeepEquals, exp)

	b := &scalar{
		0x529eec33, 0x721cf5b5, 0xc8e9c2ab, 0x7a4cf635,
		0x44a725bf, 0xeec492d9, 0x0cd77058, 0x00000002,
	}
	b.set(word(0x2aae8688))

	exp = &scalar{
		0x2aae8688, 0x721cf5b5, 0xc8e9c2ab, 0x7a4cf635,
		0x44a725bf, 0xeec492d9, 0x0cd77058, 0x00000002,
	}

	c.Assert(b, DeepEquals, exp)
}

func (s *Ed448Suite) Test_ScalarAddition(c *C) {
	a := &scalar{
		0x529eec33, 0x721cf5b5, 0xc8e9c2ab, 0x7a4cf635,
		0x44a725bf, 0xeec492d9, 0x0cd77058, 0x00000002,
	}

	b := &scalar{0x01}

	exp := &scalar{
		0x529eec34, 0x721cf5b5, 0xc8e9c2ab, 0x7a4cf635,
		0x44a725bf, 0xeec492d9, 0x0cd77058, 0x00000002,
	}

	out := new(scalar)
	out.add(a, b)

	c.Assert(out, DeepEquals, exp)
}

func (s *Ed448Suite) Test_ScalarSubtraction(c *C) {
	a := &scalar{0x0d}
	b := &scalar{0x0c}

	out := new(scalar)
	out.sub(a, b)

	exp := &scalar{0x01}

	c.Assert(out, DeepEquals, exp)
}

func (s *Ed448Suite) Test_ScalarMultiplication(c *C) {
	a := &scalar{
		0xffb823a3, 0xc96a3c35, 0x7f8ed27d, 0x087b8fb9,
		0x1d9ac30a, 0x74d65764, 0xc0be082e, 0xa8cb0ae8,
		0xa8fa552b, 0x2aae8688, 0x2c3dc273, 0x47cf8cac,
		0x3b089f07, 0x1e63e807,
	}

	b := &scalar{
		0xd8bedc42, 0x686eb329, 0xe416b899, 0x17aa6d9b,
		0x1e30b38b, 0x188c6b1a, 0xd099595b, 0xbc343bcb,
		0x1adaa0e7, 0x24e8d499, 0x8e59b308, 0x0a92de2d,
		0xcae1cb68, 0x16c5450a,
	}

	exp := &scalar{
		0xa18d010a, 0x1f5b3197, 0x994c9c2b, 0x6abd26f5,
		0x08a3a0e4, 0x36a14920, 0x74e9335f, 0x07bcd931,
		0xf2d89c1e, 0xb9036ff6, 0x203d424b, 0xfccd61b3,
		0x4ca389ed, 0x31e055c1,
	}
	a.mul(a, b)
	c.Assert(a, DeepEquals, exp)
}

func (s *Ed448Suite) Test_ScalarHalve(c *C) {
	a := &scalar{0x0c}

	out := new(scalar)
	out.halve(a)

	exp := &scalar{0x06}

	c.Assert(out, DeepEquals, exp)
}

func (s *Ed448Suite) Test_ScalarDecode(c *C) {
	buf := []byte{
		0xf5, 0x81, 0x74, 0xd5, 0x7a, 0x33, 0x72, 0x36,
		0x3c, 0x0d, 0x9f, 0xcf, 0xaa, 0x3d, 0xc1, 0x8b,
		0x1e, 0xff, 0x7e, 0x89, 0xbf, 0x76, 0x78, 0x63,
		0x65, 0x80, 0xd1, 0x7d, 0xd8, 0x4a, 0x87, 0x3b,
		0x14, 0xb9, 0xc0, 0xe1, 0x68, 0x0b, 0xbd, 0xc8,
		0x76, 0x47, 0xf3, 0xc3, 0x82, 0x90, 0x2d, 0x2f,
		0x58, 0xd2, 0x75, 0x4b, 0x39, 0xbc, 0xa8, 0x74,
	}

	exp := &scalar{
		0x2a1c3d02, 0x12f970e8, 0x41d97de7, 0x6a547b38,
		0xdaa8c88e, 0x9f299b75, 0x01075c7b, 0x3b874ad9,
		0xe1c0b914, 0xc8bd0b68, 0xc3f34776, 0x2f2d9082,
		0x4b75d258, 0x34a8bc39,
	}

	out := new(scalar)
	ok := out.decode(buf)

	c.Assert(out, DeepEquals, exp)
	c.Assert(ok, Equals, decafFalse)

	buf1 := []byte{
		0xfa, 0x77, 0xed, 0x08, 0x51, 0x91, 0xc4, 0x85,
		0x74, 0x28, 0xdd, 0xa0, 0xed, 0xbc, 0x88, 0x71,
		0xbd, 0xc3, 0x34, 0x9a, 0xce, 0xee, 0x1a, 0xab,
		0x4c, 0xa2, 0x37, 0xea, 0xb4, 0xea, 0xd2, 0x8d,
		0x25, 0xf1, 0x10, 0x86, 0xc0, 0x60, 0xeb, 0xb3,
		0xb0, 0x9a, 0xaa, 0x8a, 0x4b, 0x00, 0x9e, 0xf1,
		0x93, 0x25, 0xfe, 0x78, 0x0f, 0xdd, 0xa1, 0x3a,
	}

	exp = &scalar{
		0x08ed77fa, 0x85c49151, 0xa0dd2874, 0x7188bced,
		0x9a34c3bd, 0xab1aeece, 0xea37a24c, 0x8dd2eab4,
		0x8610f125, 0xb3eb60c0, 0x8aaa9ab0, 0xf19e004b,
		0x78fe2593, 0x3aa1dd0f,
	}

	out = new(scalar)
	ok = out.decode(buf1)

	c.Assert(out, DeepEquals, exp)
	c.Assert(ok, Equals, decafTrue)
}

func (s *Ed448Suite) Test_ScalarDecodeLong(c *C) {
	var buf []byte
	x := &scalar{}
	out := decodeLong(x, buf)

	c.Assert(out, DeepEquals, scalarZero)

	buf = []byte{
		0xf5, 0x81, 0x74, 0xd5, 0x7a, 0x33, 0x72, 0x36,
		0x3c, 0x0d, 0x9f, 0xcf, 0xaa, 0x3d, 0xc1, 0x8b,
		0x1e, 0xff, 0x7e, 0x89, 0xbf, 0x76, 0x78, 0x63,
		0x65, 0x80, 0xd1, 0x7d, 0xd8, 0x4a, 0x87, 0x3b,
		0x14, 0xb9, 0xc0, 0xe1, 0x68, 0x0b, 0xbd, 0xc8,
		0x76, 0x47, 0xf3, 0xc3, 0x82, 0x90, 0x2d, 0x2f,
		0x58, 0xd2, 0x75, 0x4b, 0x39, 0xbc, 0xa8, 0x74,
	}

	exp := &scalar{
		0x2a1c3d02, 0x12f970e8, 0x41d97de7, 0x6a547b38,
		0xdaa8c88e, 0x9f299b75, 0x01075c7b, 0x3b874ad9,
		0xe1c0b914, 0xc8bd0b68, 0xc3f34776, 0x2f2d9082,
		0x4b75d258, 0x34a8bc39,
	}

	x = scalarZero
	out = decodeLong(x, buf)

	c.Assert(out, DeepEquals, exp)

	buf = []byte{
		0xf0, 0xe4, 0x4d, 0xd4, 0x98, 0xf3, 0xad, 0x30,
		0x83, 0xe1, 0xf5, 0xfc, 0xc1, 0x44, 0xed, 0x1f,
		0xf5, 0xfb, 0x62, 0x5b, 0xa6, 0x21, 0x41, 0xa8,
		0xde, 0x2a, 0x90, 0x23, 0x13, 0xb3, 0x1a, 0xd1,
		0x41, 0x13, 0x42, 0x94, 0xdb, 0x9b, 0x0d, 0x84,
		0xec, 0x43, 0x7a, 0x51, 0x5a, 0x9b, 0x85, 0xbd,
		0xa1, 0xb1, 0x5e, 0xac, 0xeb, 0xe4, 0xa3, 0xb2,
		0x0}

	exp = &scalar{
		0x7d9d5b0a, 0xe9bc6e73, 0xe16ac2d8, 0xdd13bfdc,
		0xfdb68ed4, 0x1fa36b12, 0x29fbe30b, 0xd11ab314,
		0x94421341, 0x840d9bdb, 0x517a43ec, 0xbd859b5a,
		0xac5eb1a1, 0x32a3e4eb,
	}

	x = scalarZero
	out = decodeLong(x, buf)
	c.Assert(out, DeepEquals, exp)
}

func (s *Ed448Suite) Test_ScalarDestroy(c *C) {
	sc := &scalar{}
	copy(sc[:], ScalarQ[:])

	sc.destroy()

	c.Assert(sc, DeepEquals, scalarZero)
}

// Exported Functions
func (s *Ed448Suite) Test_NewScalar(c *C) {
	bytes := []byte{
		0x25, 0x8a, 0x52, 0x63, 0xd9, 0xf0, 0xfa, 0xad,
		0x9d, 0x50, 0x40, 0x8a, 0xf0, 0x76, 0x66, 0xe3,
		0x3d, 0xc2, 0x86, 0x1b, 0x01, 0x54, 0x18, 0xb8,
		0x1b, 0x3b, 0x76, 0xcd, 0x55, 0x18, 0xa2, 0xfd,
		0xf1, 0xf2, 0x64, 0xee, 0xae, 0xae, 0xc5, 0xe7,
		0x68, 0xa4, 0x2e, 0xde, 0x76, 0x60, 0xe6, 0x4a,
		0x51, 0x12, 0xb1, 0x35, 0x3d, 0xac, 0x04, 0x08,
	}

	exp := &scalar{
		0x63528a25, 0xadfaf0d9, 0x8a40509d, 0xe36676f0,
		0x1b86c23d, 0xb8185401, 0xcd763b1b, 0xfda21855,
		0xee64f2f1, 0xe7c5aeae, 0xde2ea468, 0x4ae66076,
		0x35b11251, 0x0804ac3d,
	}

	c.Assert(NewScalar(), DeepEquals, &scalar{})
	c.Assert(NewScalar(bytes), DeepEquals, exp)
}

func (s *Ed448Suite) Test_ExportedScalarCopy(c *C) {
	exp := &scalar{
		0xffb823a3, 0xc96a3c35, 0x7f8ed27d, 0x087b8fb9,
		0x1d9ac30a, 0x74d65764, 0xc0be082e, 0xa8cb0ae8,
		0xa8fa552b, 0x2aae8688, 0x2c3dc273, 0x47cf8cac,
		0x3b089f07, 0x1e63e807,
	}
	a := exp.Copy()
	c.Assert(a, DeepEquals, exp)
}

func (s *Ed448Suite) Test_ExportedScalarEquals(c *C) {
	a := new(scalar)
	b := a.Copy()
	c.Assert(a.Equals(b), Equals, true)
}

func (s *Ed448Suite) Test_ExportedScalarAddition(c *C) {
	a := &scalar{0x01}
	b := &scalar{0x02}
	exp := &scalar{0x03}

	out := new(scalar)
	out.Add(a, b)

	c.Assert(out, DeepEquals, exp)
}

func (s *Ed448Suite) Test_ExportedScalarSubtraction(c *C) {
	a := &scalar{0x0d}
	b := &scalar{0x0c}

	out := new(scalar)
	out.Sub(a, b)

	exp := &scalar{0x01}

	c.Assert(out, DeepEquals, exp)
}

func (s *Ed448Suite) Test_ExportedScalarMultiplication(c *C) {
	a := &scalar{
		0xffb823a3, 0xc96a3c35, 0x7f8ed27d, 0x087b8fb9,
		0x1d9ac30a, 0x74d65764, 0xc0be082e, 0xa8cb0ae8,
		0xa8fa552b, 0x2aae8688, 0x2c3dc273, 0x47cf8cac,
		0x3b089f07, 0x1e63e807,
	}

	b := &scalar{
		0xd8bedc42, 0x686eb329, 0xe416b899, 0x17aa6d9b,
		0x1e30b38b, 0x188c6b1a, 0xd099595b, 0xbc343bcb,
		0x1adaa0e7, 0x24e8d499, 0x8e59b308, 0x0a92de2d,
		0xcae1cb68, 0x16c5450a,
	}

	exp := &scalar{
		0xa18d010a, 0x1f5b3197, 0x994c9c2b, 0x6abd26f5,
		0x08a3a0e4, 0x36a14920, 0x74e9335f, 0x07bcd931,
		0xf2d89c1e, 0xb9036ff6, 0x203d424b, 0xfccd61b3,
		0x4ca389ed, 0x31e055c1,
	}
	a.Mul(a, b)
	c.Assert(a, DeepEquals, exp)
}

func (s *Ed448Suite) Test_ExportedScalarBarretDecode(c *C) {
	bytes := []byte{
		0xfa, 0x77, 0xed, 0x08, 0x51, 0x91, 0xc4, 0x85,
		0x74, 0x28, 0xdd, 0xa0, 0xed, 0xbc, 0x88, 0x71,
		0xbd, 0xc3, 0x34, 0x9a, 0xce, 0xee, 0x1a, 0xab,
		0x4c, 0xa2, 0x37, 0xea, 0xb4, 0xea, 0xd2, 0x8d,
		0x25, 0xf1, 0x10, 0x86, 0xc0, 0x60, 0xeb, 0xb3,
	}

	out := new(scalar)
	ok := out.BarretDecode(bytes)

	c.Assert(ok, ErrorMatches, "ed448: cannot decode a scalar from a byte array with a length unequal to 56")
}

func (s *Ed448Suite) Test_ExportedScalarDecode(c *C) {
	bytes := []byte{
		0xf5, 0x81, 0x74, 0xd5, 0x7a, 0x33, 0x72, 0x36,
		0x3c, 0x0d, 0x9f, 0xcf, 0xaa, 0x3d, 0xc1, 0x8b,
		0x1e, 0xff, 0x7e, 0x89, 0xbf, 0x76, 0x78, 0x63,
		0x65, 0x80, 0xd1, 0x7d, 0xd8, 0x4a, 0x87, 0x3b,
		0x14, 0xb9, 0xc0, 0xe1, 0x68, 0x0b, 0xbd, 0xc8,
		0x76, 0x47, 0xf3, 0xc3, 0x82, 0x90, 0x2d, 0x2f,
		0x58, 0xd2, 0x75, 0x4b, 0x39, 0xbc, 0xa8, 0x74,
	}

	exp := &scalar{
		0x2a1c3d02, 0x12f970e8, 0x41d97de7, 0x6a547b38,
		0xdaa8c88e, 0x9f299b75, 0x01075c7b, 0x3b874ad9,
		0xe1c0b914, 0xc8bd0b68, 0xc3f34776, 0x2f2d9082,
		0x4b75d258, 0x34a8bc39,
	}

	out := &scalar{}
	out.Decode(bytes)

	c.Assert(out, DeepEquals, exp)
}

func (s *Ed448Suite) Test_Decode_AScalarLongerThan57(c *C) {
	inp := []byte{
		0x71, 0xd6, 0x2, 0xd2, 0x13, 0x94, 0x49, 0x29,
		0x75, 0x18, 0xd1, 0xc7, 0x4e, 0x2d, 0x45, 0x44,
		0x60, 0x2a, 0xbb, 0xbb, 0x68, 0xfd, 0xb9, 0x15,
		0xd8, 0xea, 0xd9, 0xc, 0xa6, 0xd9, 0x66, 0x72,
		0x37, 0x8b, 0x2, 0xb2, 0xf1, 0xa2, 0x62, 0x9a,
		0xba, 0x56, 0xf, 0xc9, 0xb3, 0x9d, 0x8d, 0x3d,
		0xda, 0x2f, 0xe6, 0x78, 0x62, 0x27, 0x61, 0x6e,
		0xe9, 0x36, 0x13, 0xc, 0xcc, 0x34, 0x5, 0x67,
		0xd0, 0xcf, 0x76, 0x90, 0xc5, 0xf6, 0x91, 0xd7,
		0x78, 0x82, 0x44, 0xeb, 0xbe, 0xb3, 0x75, 0xc4,
		0x61, 0xee, 0x5e, 0x9c, 0x41, 0xee, 0xdc, 0xa7,
		0xbf, 0x3f, 0x91, 0x36, 0x81, 0x30, 0x12, 0x67,
		0x19, 0x68, 0x2b, 0x1c, 0x73, 0x28, 0x38, 0x5c,
		0x16, 0x72, 0xe6, 0xb9, 0x2, 0x6e, 0xe4, 0xcf,
		0x56, 0x19,
	}

	exp := []byte{
		0x12, 0x00, 0x7a, 0x28, 0x40, 0x53, 0x6a, 0xd4,
		0x89, 0xe0, 0x0c, 0x6e, 0xc7, 0xfa, 0x7a, 0xc6,
		0xe1, 0x77, 0x8e, 0x8e, 0x34, 0xb8, 0xd3, 0x5c,
		0x61, 0x84, 0x73, 0xcc, 0xb4, 0xf6, 0x38, 0x9c,
		0x6c, 0xf3, 0x2f, 0xa4, 0xca, 0x70, 0xfe, 0x2d,
		0x4f, 0xca, 0x08, 0x2b, 0x38, 0xfd, 0xc7, 0x31,
		0xf3, 0x1b, 0x6d, 0x87, 0xf5, 0x15, 0xe6, 0x1b,
	}

	sc := NewScalar(inp)
	res := sc.Encode()

	c.Assert(res, DeepEquals, exp)
}
