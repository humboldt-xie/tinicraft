package main

import "log"

var (
	tex = NewItemHub()
)

type FaceTexture [6][2]float32

func MakeFaceTexture(idx int) FaceTexture {
	const textureColums = 16
	var m = 1 / float32(textureColums)
	dx, dy := float32(idx%textureColums)*m, float32(idx/textureColums)*m
	n := float32(1 / 2048.0)
	m -= n
	return [6][2]float32{
		{dx + n, dy + n},
		{dx + m, dy + n},
		{dx + m, dy + m},
		{dx + m, dy + m},
		{dx + n, dy + m},
		{dx + n, dy + n},
	}
}

type BlockTexture struct {
	Left, Right FaceTexture
	Up, Down    FaceTexture
	Front, Back FaceTexture
}

type ItemHub struct {
	tex map[int]*BlockTexture
}

func NewItemHub() *ItemHub {
	return &ItemHub{
		tex: make(map[int]*BlockTexture),
	}
}

func (h *ItemHub) AddTexture(w int, l, r, u, d, f, b int) {
	h.tex[w] = &BlockTexture{
		Left:  MakeFaceTexture(l),
		Right: MakeFaceTexture(r),
		Up:    MakeFaceTexture(u),
		Down:  MakeFaceTexture(d),
		Front: MakeFaceTexture(f),
		Back:  MakeFaceTexture(b),
	}
}

func (h *ItemHub) Texture(w *Block) *BlockTexture {
	t, ok := h.tex[w.Type]
	if !ok {
		log.Printf("%d not found", w)
		return h.tex[0]
	}
	return t
}

func LoadTextureDesc() error {
	for w, f := range itemDesc {
		tex.AddTexture(w, f[0], f[1], f[2], f[3], f[4], f[5])
	}
	return nil
}

type ItemDesc struct {
	Texture []int
}

// w => left, right, top, bottom, front, back
var itemDesc = map[int][6]int{
	0:  {0, 0, 0, 0, 0, 0},
	1:  {16, 16, 32, 0, 16, 16},
	2:  {1, 1, 1, 1, 1, 1},
	3:  {2, 2, 2, 2, 2, 2},
	4:  {3, 3, 3, 3, 3, 3},
	5:  {20, 20, 36, 4, 20, 20},
	6:  {5, 5, 5, 5, 5, 5},
	7:  {6, 6, 6, 6, 6, 6},
	8:  {7, 7, 7, 7, 7, 7},
	9:  {24, 24, 40, 8, 24, 24},
	10: {9, 9, 9, 9, 9, 9},
	11: {10, 10, 10, 10, 10, 10},
	12: {11, 11, 11, 11, 11, 11},
	13: {12, 12, 12, 12, 12, 12},
	14: {13, 13, 13, 13, 13, 13},
	15: {14, 14, 14, 14, 14, 14},
	16: {15, 15, 15, 15, 15, 15},
	17: {48, 48, 0, 0, 48, 48},
	18: {49, 49, 0, 0, 49, 49},
	19: {50, 50, 0, 0, 50, 50},
	20: {51, 51, 0, 0, 51, 51},
	21: {52, 52, 0, 0, 52, 52},
	22: {53, 53, 0, 0, 53, 53},
	23: {54, 54, 0, 0, 54, 54},
	24: {0, 0, 0, 0, 0, 0},
	25: {0, 0, 0, 0, 0, 0},
	26: {0, 0, 0, 0, 0, 0},
	27: {0, 0, 0, 0, 0, 0},
	28: {0, 0, 0, 0, 0, 0},
	29: {0, 0, 0, 0, 0, 0},
	30: {0, 0, 0, 0, 0, 0},
	31: {0, 0, 0, 0, 0, 0},
	32: {176, 176, 176, 176, 176, 176},
	33: {177, 177, 177, 177, 177, 177},
	34: {178, 178, 178, 178, 178, 178},
	35: {179, 179, 179, 179, 179, 179},
	36: {180, 180, 180, 180, 180, 180},
	37: {181, 181, 181, 181, 181, 181},
	38: {182, 182, 182, 182, 182, 182},
	39: {183, 183, 183, 183, 183, 183},
	40: {184, 184, 184, 184, 184, 184},
	41: {185, 185, 185, 185, 185, 185},
	42: {186, 186, 186, 186, 186, 186},
	43: {187, 187, 187, 187, 187, 187},
	44: {188, 188, 188, 188, 188, 188},
	45: {189, 189, 189, 189, 189, 189},
	46: {190, 190, 190, 190, 190, 190},
	47: {191, 191, 191, 191, 191, 191},
	48: {192, 192, 192, 192, 192, 192},
	49: {193, 193, 193, 193, 193, 193},
	50: {194, 194, 194, 194, 194, 194},
	51: {195, 195, 195, 195, 195, 195},
	52: {196, 196, 196, 196, 196, 196},
	53: {197, 197, 197, 197, 197, 197},
	54: {198, 198, 198, 198, 198, 198},
	55: {199, 199, 199, 199, 199, 199},
	56: {200, 200, 200, 200, 200, 200},
	57: {201, 201, 201, 201, 201, 201},
	58: {202, 202, 202, 202, 202, 202},
	59: {203, 203, 203, 203, 203, 203},
	60: {204, 204, 204, 204, 204, 204},
	61: {205, 205, 205, 205, 205, 205},
	62: {206, 206, 206, 206, 206, 206},
	63: {207, 207, 207, 207, 207, 207},
	64: {226, 224, 241, 209, 227, 225},
}

var availableItems = []Block{
	Block{Type: 1},
	Block{Type: 2},
	Block{Type: 3},
	Block{Type: 4},
	Block{Type: 5},
	Block{Type: 6},
	Block{Type: 7},
	Block{Type: 8},
	Block{Type: 9},
	Block{Type: 10},
	Block{Type: 11},
	Block{Type: 12},
	Block{Type: 13},
	Block{Type: 14},
	Block{Type: 15},
	Block{Type: 16},
	Block{Type: 17},
	Block{Type: 18},
	Block{Type: 19},
	Block{Type: 20},
	Block{Type: 21},
	Block{Type: 22},
	Block{Type: 23},
	Block{Type: 32},
	Block{Type: 33},
	Block{Type: 34},
	Block{Type: 35},
	Block{Type: 36},
	Block{Type: 37},
	Block{Type: 38},
	Block{Type: 39},
	Block{Type: 40},
	Block{Type: 41},
	Block{Type: 42},
	Block{Type: 43},
	Block{Type: 44},
	Block{Type: 45},
	Block{Type: 46},
	Block{Type: 47},
	Block{Type: 48},
	Block{Type: 49},
	Block{Type: 50},
	Block{Type: 51},
	Block{Type: 52},
	Block{Type: 53},
	Block{Type: 54},
	Block{Type: 55},
	Block{Type: 56},
	Block{Type: 57},
	Block{Type: 58},
	Block{Type: 59},
	Block{Type: 60},
	Block{Type: 61},
	Block{Type: 62},
	Block{Type: 63},
	Block{Type: 64},
}
