package java

// This file transliterates java.util.TimSort and java.util.ComparableTimSort
// from JDK 21, the sort behind Arrays.sort, Collections.sort and List.sort
// for objects. The two JDK classes differ only in how they compare, so one
// generic port serves both. Reproducing the algorithm exactly (rather than
// using any stable sort) matters for comparators that are not consistent,
// where the result, the comparisons made, and whether an
// IllegalArgumentException is thrown all depend on the algorithm.

const (
	timSortMinMerge  = 32
	timSortMinGallop = 7
)

// timSort sorts a[lo:hi] with c, or by natural ordering if c is nil.
func timSort(a []any, lo, hi int, c Comparator) {
	if c == nil {
		timSortFunc(a, lo, hi, Comparable_CompareTo_Object)
		return
	}
	timSortFunc(a, lo, hi, c.Compare_Object_Object)
}

// timSortFunc is TimSort.sort(a, lo, hi, c, null, 0, 0).
func timSortFunc[T any](a []T, lo, hi int, cmp func(a, b T) int32) {
	nRemaining := hi - lo
	if nRemaining < 2 {
		return
	}
	if nRemaining < timSortMinMerge {
		initRunLen := countRunAndMakeAscending(a, lo, hi, cmp)
		binarySort(a, lo, hi, lo+initRunLen, cmp)
		return
	}
	ts := newTimSorter(a, cmp)
	minRun := minRunLength(nRemaining)
	for {
		runLen := countRunAndMakeAscending(a, lo, hi, cmp)
		if runLen < minRun {
			force := minRun
			if nRemaining <= minRun {
				force = nRemaining
			}
			binarySort(a, lo, lo+force, lo+runLen, cmp)
			runLen = force
		}
		ts.pushRun(lo, runLen)
		ts.mergeCollapse()
		lo += runLen
		nRemaining -= runLen
		if nRemaining == 0 {
			break
		}
	}
	ts.mergeForceCollapse()
}

type timSorter[T any] struct {
	a         []T
	cmp       func(a, b T) int32
	minGallop int
	tmp       []T
	stackSize int
	runBase   []int
	runLen    []int
}

func newTimSorter[T any](a []T, cmp func(a, b T) int32) *timSorter[T] {
	n := len(a)
	var stackLen int
	switch {
	case n < 120:
		stackLen = 5
	case n < 1542:
		stackLen = 10
	case n < 119151:
		stackLen = 24
	default:
		stackLen = 49
	}
	return &timSorter[T]{
		a:         a,
		cmp:       cmp,
		minGallop: timSortMinGallop,
		runBase:   make([]int, stackLen),
		runLen:    make([]int, stackLen),
	}
}

func binarySort[T any](a []T, lo, hi, start int, cmp func(a, b T) int32) {
	if start == lo {
		start++
	}
	for ; start < hi; start++ {
		pivot := a[start]
		left := lo
		right := start
		for left < right {
			mid := int(uint(left+right) >> 1)
			if cmp(pivot, a[mid]) < 0 {
				right = mid
			} else {
				left = mid + 1
			}
		}
		copy(a[left+1:start+1], a[left:start])
		a[left] = pivot
	}
}

func countRunAndMakeAscending[T any](a []T, lo, hi int, cmp func(a, b T) int32) int {
	runHi := lo + 1
	if runHi == hi {
		return 1
	}
	if cmp(a[runHi], a[lo]) < 0 {
		runHi++
		for runHi < hi && cmp(a[runHi], a[runHi-1]) < 0 {
			runHi++
		}
		reverseRange(a, lo, runHi)
	} else {
		runHi++
		for runHi < hi && cmp(a[runHi], a[runHi-1]) >= 0 {
			runHi++
		}
	}
	return runHi - lo
}

func reverseRange[T any](a []T, lo, hi int) {
	hi--
	for lo < hi {
		a[lo], a[hi] = a[hi], a[lo]
		lo++
		hi--
	}
}

func minRunLength(n int) int {
	r := 0
	for n >= timSortMinMerge {
		r |= n & 1
		n >>= 1
	}
	return n + r
}

func (ts *timSorter[T]) pushRun(runBase, runLen int) {
	ts.runBase[ts.stackSize] = runBase
	ts.runLen[ts.stackSize] = runLen
	ts.stackSize++
}

func (ts *timSorter[T]) mergeCollapse() {
	for ts.stackSize > 1 {
		n := ts.stackSize - 2
		runLen := ts.runLen
		if n > 0 && runLen[n-1] <= runLen[n]+runLen[n+1] || n > 1 && runLen[n-2] <= runLen[n]+runLen[n-1] {
			if runLen[n-1] < runLen[n+1] {
				n--
			}
		} else if n < 0 || runLen[n] > runLen[n+1] {
			break
		}
		ts.mergeAt(n)
	}
}

func (ts *timSorter[T]) mergeForceCollapse() {
	for ts.stackSize > 1 {
		n := ts.stackSize - 2
		if n > 0 && ts.runLen[n-1] < ts.runLen[n+1] {
			n--
		}
		ts.mergeAt(n)
	}
}

func (ts *timSorter[T]) mergeAt(i int) {
	base1 := ts.runBase[i]
	len1 := ts.runLen[i]
	base2 := ts.runBase[i+1]
	len2 := ts.runLen[i+1]
	ts.runLen[i] = len1 + len2
	if i == ts.stackSize-3 {
		ts.runBase[i+1] = ts.runBase[i+2]
		ts.runLen[i+1] = ts.runLen[i+2]
	}
	ts.stackSize--

	k := gallopRight(ts.a[base2], ts.a, base1, len1, 0, ts.cmp)
	base1 += k
	len1 -= k
	if len1 == 0 {
		return
	}
	len2 = gallopLeft(ts.a[base1+len1-1], ts.a, base2, len2, len2-1, ts.cmp)
	if len2 == 0 {
		return
	}
	if len1 <= len2 {
		ts.mergeLo(base1, len1, base2, len2)
	} else {
		ts.mergeHi(base1, len1, base2, len2)
	}
}

func gallopLeft[T any](key T, a []T, base, length, hint int, cmp func(a, b T) int32) int {
	lastOfs := 0
	ofs := 1
	if cmp(key, a[base+hint]) > 0 {
		maxOfs := length - hint
		for ofs < maxOfs && cmp(key, a[base+hint+ofs]) > 0 {
			lastOfs = ofs
			ofs = (ofs << 1) + 1
			if ofs <= 0 {
				ofs = maxOfs
			}
		}
		if ofs > maxOfs {
			ofs = maxOfs
		}
		lastOfs += hint
		ofs += hint
	} else {
		maxOfs := hint + 1
		for ofs < maxOfs && cmp(key, a[base+hint-ofs]) <= 0 {
			lastOfs = ofs
			ofs = (ofs << 1) + 1
			if ofs <= 0 {
				ofs = maxOfs
			}
		}
		if ofs > maxOfs {
			ofs = maxOfs
		}
		tmp := lastOfs
		lastOfs = hint - ofs
		ofs = hint - tmp
	}
	lastOfs++
	for lastOfs < ofs {
		m := lastOfs + int(uint(ofs-lastOfs)>>1)
		if cmp(key, a[base+m]) > 0 {
			lastOfs = m + 1
		} else {
			ofs = m
		}
	}
	return ofs
}

func gallopRight[T any](key T, a []T, base, length, hint int, cmp func(a, b T) int32) int {
	ofs := 1
	lastOfs := 0
	if cmp(key, a[base+hint]) < 0 {
		maxOfs := hint + 1
		for ofs < maxOfs && cmp(key, a[base+hint-ofs]) < 0 {
			lastOfs = ofs
			ofs = (ofs << 1) + 1
			if ofs <= 0 {
				ofs = maxOfs
			}
		}
		if ofs > maxOfs {
			ofs = maxOfs
		}
		tmp := lastOfs
		lastOfs = hint - ofs
		ofs = hint - tmp
	} else {
		maxOfs := length - hint
		for ofs < maxOfs && cmp(key, a[base+hint+ofs]) >= 0 {
			lastOfs = ofs
			ofs = (ofs << 1) + 1
			if ofs <= 0 {
				ofs = maxOfs
			}
		}
		if ofs > maxOfs {
			ofs = maxOfs
		}
		lastOfs += hint
		ofs += hint
	}
	lastOfs++
	for lastOfs < ofs {
		m := lastOfs + int(uint(ofs-lastOfs)>>1)
		if cmp(key, a[base+m]) < 0 {
			ofs = m
		} else {
			lastOfs = m + 1
		}
	}
	return ofs
}

func (ts *timSorter[T]) ensureCapacity(n int) []T {
	if len(ts.tmp) < n {
		ts.tmp = make([]T, n)
	}
	return ts.tmp
}

func timSortContractViolation() any {
	return IllegalArgumentException_new_String(Lit("Comparison method violates its general contract!"))
}

func (ts *timSorter[T]) mergeLo(base1, len1, base2, len2 int) {
	a := ts.a
	cmp := ts.cmp
	tmp := ts.ensureCapacity(len1)
	cursor1 := 0
	cursor2 := base2
	dest := base1
	copy(tmp[cursor1:cursor1+len1], a[base1:base1+len1])

	a[dest] = a[cursor2]
	dest++
	cursor2++
	len2--
	if len2 == 0 {
		copy(a[dest:dest+len1], tmp[cursor1:cursor1+len1])
		return
	}
	if len1 == 1 {
		copy(a[dest:dest+len2], a[cursor2:cursor2+len2])
		a[dest+len2] = tmp[cursor1]
		return
	}

	minGallop := ts.minGallop
outer:
	for {
		count1 := 0
		count2 := 0
		for {
			if cmp(a[cursor2], tmp[cursor1]) < 0 {
				a[dest] = a[cursor2]
				dest++
				cursor2++
				count2++
				count1 = 0
				len2--
				if len2 == 0 {
					break outer
				}
			} else {
				a[dest] = tmp[cursor1]
				dest++
				cursor1++
				count1++
				count2 = 0
				len1--
				if len1 == 1 {
					break outer
				}
			}
			if (count1 | count2) >= minGallop {
				break
			}
		}
		for {
			count1 = gallopRight(a[cursor2], tmp, cursor1, len1, 0, cmp)
			if count1 != 0 {
				copy(a[dest:dest+count1], tmp[cursor1:cursor1+count1])
				dest += count1
				cursor1 += count1
				len1 -= count1
				if len1 <= 1 {
					break outer
				}
			}
			a[dest] = a[cursor2]
			dest++
			cursor2++
			len2--
			if len2 == 0 {
				break outer
			}
			count2 = gallopLeft(tmp[cursor1], a, cursor2, len2, 0, cmp)
			if count2 != 0 {
				copy(a[dest:dest+count2], a[cursor2:cursor2+count2])
				dest += count2
				cursor2 += count2
				len2 -= count2
				if len2 == 0 {
					break outer
				}
			}
			a[dest] = tmp[cursor1]
			dest++
			cursor1++
			len1--
			if len1 == 1 {
				break outer
			}
			minGallop--
			if !(count1 >= timSortMinGallop || count2 >= timSortMinGallop) {
				break
			}
		}
		if minGallop < 0 {
			minGallop = 0
		}
		minGallop += 2
	}
	if minGallop < 1 {
		minGallop = 1
	}
	ts.minGallop = minGallop

	switch {
	case len1 == 1:
		copy(a[dest:dest+len2], a[cursor2:cursor2+len2])
		a[dest+len2] = tmp[cursor1]
	case len1 == 0:
		panic(timSortContractViolation())
	default:
		copy(a[dest:dest+len1], tmp[cursor1:cursor1+len1])
	}
}

func (ts *timSorter[T]) mergeHi(base1, len1, base2, len2 int) {
	a := ts.a
	cmp := ts.cmp
	tmp := ts.ensureCapacity(len2)
	tmpBase := 0
	copy(tmp[tmpBase:tmpBase+len2], a[base2:base2+len2])

	cursor1 := base1 + len1 - 1
	cursor2 := tmpBase + len2 - 1
	dest := base2 + len2 - 1

	a[dest] = a[cursor1]
	dest--
	cursor1--
	len1--
	if len1 == 0 {
		copy(a[dest-(len2-1):dest+1], tmp[tmpBase:tmpBase+len2])
		return
	}
	if len2 == 1 {
		dest -= len1
		cursor1 -= len1
		copy(a[dest+1:dest+1+len1], a[cursor1+1:cursor1+1+len1])
		a[dest] = tmp[cursor2]
		return
	}

	minGallop := ts.minGallop
outer:
	for {
		count1 := 0
		count2 := 0
		for {
			if cmp(tmp[cursor2], a[cursor1]) < 0 {
				a[dest] = a[cursor1]
				dest--
				cursor1--
				count1++
				count2 = 0
				len1--
				if len1 == 0 {
					break outer
				}
			} else {
				a[dest] = tmp[cursor2]
				dest--
				cursor2--
				count2++
				count1 = 0
				len2--
				if len2 == 1 {
					break outer
				}
			}
			if (count1 | count2) >= minGallop {
				break
			}
		}
		for {
			count1 = len1 - gallopRight(tmp[cursor2], a, base1, len1, len1-1, cmp)
			if count1 != 0 {
				dest -= count1
				cursor1 -= count1
				len1 -= count1
				copy(a[dest+1:dest+1+count1], a[cursor1+1:cursor1+1+count1])
				if len1 == 0 {
					break outer
				}
			}
			a[dest] = tmp[cursor2]
			dest--
			cursor2--
			len2--
			if len2 == 1 {
				break outer
			}
			count2 = len2 - gallopLeft(a[cursor1], tmp, tmpBase, len2, len2-1, cmp)
			if count2 != 0 {
				dest -= count2
				cursor2 -= count2
				len2 -= count2
				copy(a[dest+1:dest+1+count2], tmp[cursor2+1:cursor2+1+count2])
				if len2 <= 1 {
					break outer
				}
			}
			a[dest] = a[cursor1]
			dest--
			cursor1--
			len1--
			if len1 == 0 {
				break outer
			}
			minGallop--
			if !(count1 >= timSortMinGallop || count2 >= timSortMinGallop) {
				break
			}
		}
		if minGallop < 0 {
			minGallop = 0
		}
		minGallop += 2
	}
	if minGallop < 1 {
		minGallop = 1
	}
	ts.minGallop = minGallop

	switch {
	case len2 == 1:
		dest -= len1
		cursor1 -= len1
		copy(a[dest+1:dest+1+len1], a[cursor1+1:cursor1+1+len1])
		a[dest] = tmp[cursor2]
	case len2 == 0:
		panic(timSortContractViolation())
	default:
		copy(a[dest-(len2-1):dest+1], tmp[tmpBase:tmpBase+len2])
	}
}

// listSortDefault is List.sort's default implementation.
func listSortDefault(l List, c Comparator) {
	a := l.ToArray()
	timSort(a, 0, len(a), c)
	it := l.ListIterator()
	for _, e := range a {
		it.Next()
		it.Set_Object(e)
	}
}
