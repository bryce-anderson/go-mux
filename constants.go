package mux

////////////////////////////
// Mux Constants
////////////////////////////

const MaxStreamId = 0x007fffff	// 23 bits
const FragmentMask = 1 << 23	// 24th bit signals fragment

const TdispatchTpe = 2
const RdispatchTpe = -2

const TpingTpe = 65
const RpingTpe = -65

const TinitTpe = 68
const RinitTpe = -68

const BadRerrTpe = 127	// Old implementation fluke... Two's complement and all...
const RerrTpe = -128
