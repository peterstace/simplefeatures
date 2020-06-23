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
		{pop: 6, want: 2189616554920753830},
		{pop: 7, want: 18175851834761875554},
		{pop: 8, want: 12628255421337798194},
		{pop: 9, want: 2265426897104517421},
		{pop: 10, want: 3134134291419311046},
		{pop: 11, want: 5689804115723382764},
		{pop: 12, want: 9694228338494197849},
		{pop: 13, want: 13103729161659517582},
		{pop: 14, want: 10141211141433144241},
		{pop: 15, want: 4266143990412453129},
		{pop: 16, want: 3347339997226441897},
		{pop: 17, want: 492585592469164258},
		{pop: 18, want: 8536390920161296879},
		{pop: 19, want: 2284121401319000681},
		{pop: 20, want: 5825941524697753701},
		{pop: 21, want: 3971074051373273461},
		{pop: 22, want: 13500866762608516470},
		{pop: 23, want: 3180980945022923615},
		{pop: 24, want: 10702125374746869609},
		{pop: 25, want: 8532266638239458606},
		{pop: 26, want: 5405728551686151941},
		{pop: 27, want: 18047497124313027793},
		{pop: 29, want: 15046331184773809950},
		{pop: 30, want: 12070730015462108450},
		{pop: 31, want: 9277304203210608327},
		{pop: 32, want: 14451055237249558456},
		{pop: 33, want: 16336008616807849411},
		{pop: 34, want: 10554984683477153544},
		{pop: 35, want: 15030612586458235427},
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
		{pop: 64, want: 3049993816560312456},
		{pop: 65, want: 12922461014132491866},
		{pop: 66, want: 2376426109815800291},
		{pop: 67, want: 3905000612714128291},
		{pop: 68, want: 4164893501225923199},
		{pop: 69, want: 8875826777212825658},
		{pop: 70, want: 12263084737349429115},
		{pop: 71, want: 5276725395669972196},
		{pop: 72, want: 5231863365026697049},
		{pop: 73, want: 8306526314392317945},
		{pop: 74, want: 10063797265031961355},
		{pop: 75, want: 16022224486892809308},
		{pop: 76, want: 13169521769841471129},
		{pop: 77, want: 3865916890536153926},
		{pop: 78, want: 10993591685083419424},
		{pop: 79, want: 18295640991331248034},
		{pop: 80, want: 6621692530838235011},
		{pop: 81, want: 9527196386043363147},
		{pop: 82, want: 7735880669638773182},
		{pop: 83, want: 13556927587036161272},
		{pop: 84, want: 4533299431683956279},
		{pop: 85, want: 15225629695082037476},
		{pop: 86, want: 6927306809264689187},
		{pop: 87, want: 17973684974838225798},
		{pop: 88, want: 2149375707361860021},
		{pop: 89, want: 15287715347702375529},
		{pop: 90, want: 13008798289484292663},
		{pop: 91, want: 4708144056942851668},
		{pop: 92, want: 138305150330276228},
		{pop: 93, want: 7295445704486788277},
		{pop: 94, want: 9321653242993893576},
		{pop: 95, want: 12618751843059854686},
		{pop: 96, want: 12924932719708780673},
		{pop: 97, want: 15455247500650356336},
		{pop: 98, want: 15405425168311053431},
		{pop: 99, want: 3106805947102690509},
		{pop: 100, want: 14966132007611414076},
		{pop: 1000, want: 6535743965441116214},
		{pop: 10_000, want: 7407227297016950893},
		{pop: 100_000, want: 14387712569501511190},
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
