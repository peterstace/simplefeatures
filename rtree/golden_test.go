package rtree

import (
	"fmt"
	"hash/crc64"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestBulkLoadGolden(t *testing.T) {
	for _, tt := range []struct {
		pop  int
		want uint64
	}{
		// Test data is 'golden'. We don't really care what the values are,
		// just that they remain stable over time. If they unexpectedly change,
		// then that's an indication that the structure of the bulkloaded tree
		// has changed (which may or may not be ok depending on the reason for
		// the change).
		{pop: 1, want: 4796333603149578240},
		{pop: 2, want: 4860108095059132416},
		{pop: 3, want: 4729504678986907648},
		{pop: 4, want: 4616912695452668560},
		{pop: 5, want: 4329441588449081019},
		{pop: 6, want: 8136983393899489447},
		{pop: 7, want: 15891291518287925241},
		{pop: 8, want: 9467914180847604717},
		{pop: 9, want: 2265426897104517421},
		{pop: 10, want: 3134134291419311046},
		{pop: 11, want: 5689804115723382764},
		{pop: 12, want: 9694228338494197849},
		{pop: 13, want: 13103729161659517582},
		{pop: 14, want: 10141211141433144241},
		{pop: 15, want: 4266143990412453129},
		{pop: 16, want: 3347339997226441897},
		{pop: 17, want: 1249590671566842103},
		{pop: 18, want: 1777094726460416551},
		{pop: 19, want: 3893977871631166558},
		{pop: 20, want: 5825941524697753701},
		{pop: 21, want: 11897939877783289987},
		{pop: 22, want: 11334843141664092413},
		{pop: 23, want: 11529801659867238957},
		{pop: 24, want: 7138758598502500264},
		{pop: 25, want: 14977117916620236013},
		{pop: 26, want: 7406063316557102263},
		{pop: 27, want: 7322544965613465078},
		{pop: 28, want: 7079409464866337190},
		{pop: 29, want: 75458395813755652},
		{pop: 30, want: 5106397057557886046},
		{pop: 31, want: 10032304007843990088},
		{pop: 32, want: 13308338555103055184},
		{pop: 33, want: 9938999414147363299},
		{pop: 34, want: 4514620220884270644},
		{pop: 35, want: 7539498635742650207},
		{pop: 36, want: 3414215425993200344},
		{pop: 37, want: 13167791222048855311},
		{pop: 38, want: 13792063080954478823},
		{pop: 39, want: 12543309934895999977},
		{pop: 40, want: 17188216630467953360},
		{pop: 41, want: 11459107173723650557},
		{pop: 42, want: 16108287302821613129},
		{pop: 43, want: 7826836058168921242},
		{pop: 44, want: 11221410816658499022},
		{pop: 45, want: 6281263094284742349},
		{pop: 46, want: 2153554965040204714},
		{pop: 47, want: 9891268565429707338},
		{pop: 48, want: 16552527092936270116},
		{pop: 49, want: 13641819854152992915},
		{pop: 50, want: 1060060456073594678},
		{pop: 51, want: 18270188283513622870},
		{pop: 52, want: 2119641369824367888},
		{pop: 53, want: 17743624411093699880},
		{pop: 54, want: 12933898159734605795},
		{pop: 55, want: 14535702187224943217},
		{pop: 56, want: 9776779229032027286},
		{pop: 57, want: 5776027755553856143},
		{pop: 58, want: 14509461278622831435},
		{pop: 59, want: 10186037722718299438},
		{pop: 60, want: 13836256746924334355},
		{pop: 61, want: 6372596478443342396},
		{pop: 62, want: 16281786708995097100},
		{pop: 63, want: 17132417846997343708},
		{pop: 64, want: 17793088422319323540},
		{pop: 65, want: 17425450922685778469},
		{pop: 66, want: 9939071655524841645},
		{pop: 67, want: 4127303398172896594},
		{pop: 68, want: 15299039166796030931},
		{pop: 69, want: 2166249301626364743},
		{pop: 70, want: 5173450520559829397},
		{pop: 71, want: 11959310751289426798},
		{pop: 72, want: 8877585929533451102},
		{pop: 73, want: 11981109536826821080},
		{pop: 74, want: 12949585872757370463},
		{pop: 75, want: 4503431580146526420},
		{pop: 76, want: 14028848284481126201},
		{pop: 77, want: 952734170165351842},
		{pop: 78, want: 1380858960473413350},
		{pop: 79, want: 8824789226657288571},
		{pop: 80, want: 7186870586647801392},
		{pop: 81, want: 16627968457730555011},
		{pop: 82, want: 15325368732487727811},
		{pop: 83, want: 12721099594672408416},
		{pop: 84, want: 5899861281714184115},
		{pop: 85, want: 3777099821639220516},
		{pop: 86, want: 11533092596164188080},
		{pop: 87, want: 15315320731847037109},
		{pop: 88, want: 3734124985378196973},
		{pop: 89, want: 16907768322889781771},
		{pop: 90, want: 2704228504945966526},
		{pop: 91, want: 2146069266454526101},
		{pop: 92, want: 159223186453704597},
		{pop: 93, want: 17757549057512864884},
		{pop: 94, want: 3060725400394765949},
		{pop: 95, want: 2431629897405091668},
		{pop: 96, want: 3728421066048302920},
		{pop: 97, want: 11211775731199352343},
		{pop: 98, want: 9002510079391438661},
		{pop: 99, want: 12039875665248478398},
		{pop: 100, want: 12194808840654274557},
		{pop: 1000, want: 9991940504894338516},
		{pop: 10_000, want: 16066516270726112266},
		{pop: 100_000, want: 15249051974644088932},
	} {
		t.Run(fmt.Sprintf("n=%d", tt.pop), func(t *testing.T) {
			rnd := rand.New(rand.NewSource(0))
			rt, _ := testBulkLoad(rnd, tt.pop, 0.9, 0.1)
			got := checksum(rt.root)
			if got != tt.want {
				t.Errorf("got=%d want=%d", got, tt.want)
			}
		})
	}
}

func checksum(n *node) uint64 {
	var entries []string
	for i := 0; i < n.numEntries; i++ {
		var entry string
		if n.isLeaf {
			entry = strconv.Itoa(n.entries[i].recordID)
		} else {
			entry = strconv.FormatUint(checksum(n.entries[i].child), 10)
		}
		entries = append(entries, entry)
	}
	sort.Strings(entries)
	return crc64.Checksum([]byte(strings.Join(entries, ",")), crc64.MakeTable(crc64.ISO))
}
