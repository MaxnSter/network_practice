package topk

/*
 多机找出最频繁出现的数据
 分三种情况:
 第一是输入文件比较小，能完全放入内存；
 第二是输入文件比较大，不能一次性都放入内存；
 第三是输入文件分布在多台机器上，这需要用到网络编程。
 文章见: https://blog.csdn.net/solstice/article/details/8497475
 视频见: http://boolan.com/course/section/1000002183
*/
