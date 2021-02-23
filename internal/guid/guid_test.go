package guid

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/convert"
	"project/internal/patch/monkey"
	"project/internal/random"
	"project/internal/testsuite"
)

func TestGUID_Write(t *testing.T) {
	expected := bytes.Repeat([]byte{1}, Size)
	guid := GUID{}

	err := guid.Write(expected)
	require.NoError(t, err)
	require.Equal(t, expected, guid[:])

	// invalid slice size
	err = guid.Write(bytes.Repeat([]byte{1}, Size-1))
	require.Error(t, err)
}

func TestGUID_String(t *testing.T) {
	guid := GUID{}
	copy(guid[Size/2:], bytes.Repeat([]byte{10}, Size/2))

	buf := bytes.Buffer{}
	buf.WriteString(strings.Repeat("00", Size/2))
	buf.WriteString(strings.Repeat("0A", Size/2))

	require.Equal(t, buf.String(), guid.String())
}

func TestGUID_Print(t *testing.T) {
	guid := GUID{}
	copy(guid[Size/2:], bytes.Repeat([]byte{10}, Size/2))

	buf := bytes.Buffer{}
	buf.WriteString("GUID: ")
	buf.WriteString(strings.Repeat("00", Size/2))
	buf.WriteString(strings.Repeat("0A", Size/2))

	require.Equal(t, buf.String(), guid.Print())
}

func TestGUID_Timestamp(t *testing.T) {
	now := time.Now().Unix()
	guid := GUID{}
	copy(guid[20:28], convert.BEInt64ToBytes(now))

	require.Equal(t, now, guid.Timestamp())
}

func TestGUID_IsZero(t *testing.T) {
	guid := GUID{}
	require.True(t, guid.IsZero())
	guid[0] = 1
	require.False(t, guid.IsZero())
}

func TestGUID_MarshalJSON(t *testing.T) {
	guid := GUID{}
	data := bytes.Repeat([]byte{10}, Size)
	copy(guid[:], data)

	data, err := guid.MarshalJSON()
	require.NoError(t, err)

	// "0101...0101"
	expected := fmt.Sprintf("\"%s\"", strings.Repeat("0A", Size))
	require.Equal(t, expected, string(data))
}

func TestGUID_UnmarshalJSON(t *testing.T) {
	data := []byte(fmt.Sprintf("\"%s\"", strings.Repeat("0A", Size)))
	guid := GUID{}

	err := guid.UnmarshalJSON(data)
	require.NoError(t, err)

	expected := bytes.Repeat([]byte{10}, Size)
	require.Equal(t, expected, guid[:])

	// invalid size
	err = guid.UnmarshalJSON(nil)
	require.Error(t, err)
}

func TestGUID_JSONUnmarshal(t *testing.T) {
	const format = `{"data": "%s"}`
	jsonData := []byte(fmt.Sprintf(format, strings.Repeat("01", Size)))

	testdata := struct {
		Data GUID `json:"data"`
	}{}
	err := json.Unmarshal(jsonData, &testdata)
	require.NoError(t, err)

	expected := bytes.Repeat([]byte{1}, Size)
	require.Equal(t, expected, testdata.Data[:])

	jsonData, err = json.Marshal(testdata)
	require.NoError(t, err)
	fmt.Println(string(jsonData))
}

func testPrintGUID(t testing.TB, guid *GUID) {
	t.Log("------------------------------GUID------------------------------")
	t.Log(guid)
	t.Log(guid.Print())
	t.Log(guid[:])
	t.Log("----------------------------------------------------------------")
}

func TestGenerator(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("with nil now", func(t *testing.T) {
		g := NewGenerator(16, nil)

		for i := 0; i < 3; i++ {
			guid := g.Get()
			testPrintGUID(t, guid)
			g.Put(guid)
		}

		g.Close()

		testsuite.IsDestroyed(t, g)
	})

	t.Run("with time.Now", func(t *testing.T) {
		g := NewGenerator(16, time.Now)

		for i := 0; i < 3; i++ {
			guid := g.Get()
			testPrintGUID(t, guid)
			g.Put(guid)
		}

		g.Close()

		testsuite.IsDestroyed(t, g)
	})

	t.Run("zero size", func(t *testing.T) {
		g := NewGenerator(0, time.Now)

		for i := 0; i < 3; i++ {
			guid := g.Get()
			testPrintGUID(t, guid)
			g.Put(guid)
		}

		// twice
		g.Close()
		g.Close()

		testsuite.IsDestroyed(t, g)
	})

	t.Run("Get() after Close()", func(t *testing.T) {
		g := NewGenerator(2, time.Now)
		time.Sleep(time.Second)
		g.Close()

		for i := 0; i < 3; i++ {
			guid := g.Get()
			testPrintGUID(t, guid)
			g.Put(guid)
		}

		testsuite.IsDestroyed(t, g)
	})

	t.Run("panic in generateLoop()", func(t *testing.T) {
		var pg *monkey.PatchGuard
		patch := func(interface{}, []byte, uint32) {
			pg.Unpatch()
			panic(monkey.Panic)
		}
		pg = monkey.PatchInstanceMethod(binary.BigEndian, "PutUint32", patch)
		defer pg.Unpatch()

		g := NewGenerator(0, time.Now)

		for i := 0; i < 3; i++ {
			guid := g.Get()
			testPrintGUID(t, guid)
			g.Put(guid)
		}

		g.Close()

		testsuite.IsDestroyed(t, g)
	})
}

func TestGenerator_Get_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("part", func(t *testing.T) {
		g := NewGenerator(512, nil)

		get := func() {
			guid := g.Get()
			require.False(t, guid.IsZero())
			g.Put(guid)
		}
		testsuite.RunParallel(100, nil, nil, get, get)

		g.Close()

		testsuite.IsDestroyed(t, g)
	})

	t.Run("whole", func(t *testing.T) {
		var g *Generator

		init := func() {
			g = NewGenerator(512, nil)
		}
		get := func() {
			guid := g.Get()
			require.False(t, guid.IsZero())
			g.Put(guid)
		}
		cleanup := func() {
			g.Close()
		}
		testsuite.RunParallel(100, init, cleanup, get, get)

		testsuite.IsDestroyed(t, g)
	})
}

func TestGenerator_Close_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("part", func(t *testing.T) {
		g := NewGenerator(512, nil)

		close1 := func() {
			g.Close()
		}
		testsuite.RunParallel(100, nil, nil, close1, close1)

		testsuite.IsDestroyed(t, g)
	})

	t.Run("whole", func(t *testing.T) {
		var g *Generator

		init := func() {
			g = NewGenerator(512, nil)
		}
		close1 := func() {
			g.Close()
		}
		testsuite.RunParallel(100, init, nil, close1, close1)

		testsuite.IsDestroyed(t, g)
	})
}

func TestGenerator_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("part", func(t *testing.T) {
		g := NewGenerator(512, nil)

		get := func() {
			g.Put(g.Get())
		}
		close1 := func() {
			g.Close()
		}
		cleanup := func() {
			g.Close()
		}
		testsuite.RunParallel(100, nil, cleanup, get, get, close1, close1)

		testsuite.IsDestroyed(t, g)
	})

	t.Run("whole", func(t *testing.T) {
		var g *Generator

		init := func() {
			g = NewGenerator(512, nil)
		}
		get := func() {
			g.Put(g.Get())
		}
		close1 := func() {
			g.Close()
		}
		cleanup := func() {
			g.Close()
		}
		testsuite.RunParallel(100, init, cleanup, get, get, close1, close1)

		testsuite.IsDestroyed(t, g)
	})
}

func BenchmarkGUID_IsZero(b *testing.B) {
	guid := GUID{}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if !guid.IsZero() {
			b.Fatal("guid is not zero")
		}
	}

	b.StopTimer()
}

func BenchmarkGenerator_Get_WithoutCache(b *testing.B) {
	gm := testsuite.MarkGoroutines(b)
	defer gm.Compare()

	g := NewGenerator(512, nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		g.Get()
	}

	b.StopTimer()

	g.Close()

	testsuite.IsDestroyed(b, g)
}

func BenchmarkGenerator_Get_WithCache(b *testing.B) {
	gm := testsuite.MarkGoroutines(b)
	defer gm.Compare()

	g := NewGenerator(512, nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		g.Put(g.Get())
	}

	b.StopTimer()

	g.Close()

	testsuite.IsDestroyed(b, g)
}

func BenchmarkGUIDForMapKey(b *testing.B) {
	gm := testsuite.MarkGoroutines(b)
	defer gm.Compare()

	rand := random.NewRand()
	key := make([]GUID, b.N)
	for i := 0; i < b.N; i++ {
		b := rand.Bytes(Size)
		copy(key[i][:], b)
	}
	m := make(map[GUID]int)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m[key[i]] = i
	}

	b.StopTimer()
}
