/*
NOTE:
Median defined as:

[1, 2, 3, 4, 5] => 3
[1, 2, 3, 4, 5, 6] => 3, not 3.5
*/
package kth

import (
	"fmt"

	"github.com/MaxnSter/gnet/logger"
)

//不过要求得到3.5这个结果也很简单 (FindKth(3)+FindKth(4))/2就好
// 传入一个猜测的数字,返回数组中比这个数小,与这个数相等的元素的个数
type Search func(guess int) (smaller, same int)

// 寻找第K大的数
func FindKth(search Search, k, count, min, max int) (int, bool) {
	if k > count {
		logger.Errorln("Algorithm failed, k max than count")
		return 0, false
	}
	// 首先随便猜测一个数字
	steps, guess, succeed := 0, max, false
	for min <= max {
		smaller, same := search(guess)
		fmt.Printf("guess = %d, smaller = %d, same = %d, min = %d, max= %d\n",
			guess, smaller, same, min, max)

		steps++
		if steps > 100 {
			//除非数组长度:2^100
			logger.Errorln("Algorithm failed, too many steps")
			break
		}

		//若guess正好是第k大的数,且guess唯一
		//此时smaller+same==k
		if smaller < k && smaller+same >= k {
			succeed = true
			break
		}

		if smaller+same < k {
			//猜测的数偏小了
			min = guess + 1
		} else if smaller >= k {
			//猜测的数偏大了
			max = guess
		} else {
			logger.Errorln("Algorithm failed, no improvement")
			break
		}
		guess = min + (max-min)/2
	}

	return guess, succeed
}
