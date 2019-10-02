package bufferpool

const (
	BUF_MIN_LEN  = 1024
	BUF_MAX_LEN  = 4 * 1024 * 1024
	BUF_RSV_LEN  = 1
	MEM_ARR_SIZE = 6
	MAX_SIZE     = 8192
)

// bufferpool =
//     256 * 1000 * 4 = 1M
//     512 * 1000 * 2 = 1M
//    1024 * 1000     = 1M
//    2048 *  512     = 1M
//    4096 *  256     = 1M
//    8192 *  128     = 1M
var memSize [MEM_ARR_SIZE]int = [MEM_ARR_SIZE]int{256, 512, 1024, 2048, 4096, 8192}
var memCnt [MEM_ARR_SIZE]int = [MEM_ARR_SIZE]int{1000 * 4, 1000 * 2, 1000, 512, 256, 128}
