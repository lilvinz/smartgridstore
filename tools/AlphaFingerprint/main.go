package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	btrdb "github.com/BTrDB/btrdb"
	"github.com/pborman/uuid"
)

var btrdbconns []*btrdb.BTrDB
var bulkstreams []*btrdb.Stream

const parallelism = 200
const connections = 6
const bulkDays = 1
const bulkFrequency = 120
const bulkBatchSize = 100000
const pqmHours = 6
const pqmFrequency = 120
const pqmBatchSize = 120

var gsetcode int

func main() {

	initialize()
	cfg := fmt.Sprintf(`"parallelism":%d, "connections":%d, "bulkDays":%d, "bulkFrequency":%d, "bulkBatchSize":%d, "pqmHours":%d, "pqmFrequency":%d,"pqmBatchSize":%d`,
		parallelism, connections, bulkDays, bulkFrequency, bulkBatchSize, pqmHours, pqmFrequency, pqmBatchSize)
	//Test tree insert throughput:
	//  Time Bulk insert 12 streams with 1 month of data each, using 12 workers

	then := time.Now()
	bulkInsert()
	bulkDeltaInsert := time.Since(then)
	bulkTotalpoints := parallelism * bulkDays * 24 * 60 * 60 * bulkFrequency
	bulkMPS := float64(bulkTotalpoints) / (float64(bulkDeltaInsert) / 1e9) / 1e6
	fmt.Printf("Bulk direct insert took: %s (%.3f MP/s)\n", bulkDeltaInsert, bulkMPS)
	fmt.Printf(`>>> {%s, "op":"bulkinsert","time":%d,"points":%d,"mps":%f}\n`, cfg, bulkDeltaInsert, bulkTotalpoints, bulkMPS)

	then = time.Now()
	pqmInsert()
	pqmDeltaInsert := time.Since(then)
	pqmTotalPoints := parallelism * pqmHours * 60 * 60 * pqmFrequency
	pqmMPS := float64(pqmTotalPoints) / (float64(pqmDeltaInsert) / 1e9) / 1e6
	fmt.Printf("PQM insert took: %s (%.3f MP/s)\n", pqmDeltaInsert, pqmMPS)
	fmt.Printf(`>>> {%s, "op","pqminsert","time":%d,"points":%d,"mps":%f}\n`, cfg, pqmDeltaInsert, pqmTotalPoints, pqmMPS)

	then = time.Now()
	sequentialLowLevel()
	deltaSequentialLowLevel := time.Since(then)
	sequentialLowLevelMPS := float64(bulkTotalpoints) / (float64(deltaSequentialLowLevel) / 1e9) / 1e6
	fmt.Printf("WARM CACHE Sequential low level took: %s (%.3f MP/s)\n", deltaSequentialLowLevel, sequentialLowLevelMPS)
	fmt.Printf(`>>> {%s, "op":"read","level":"low","pattern":"sequential","cache":"warm","time":%d,"points":%d,"mps":%f}`+"\n", cfg, deltaSequentialLowLevel, bulkTotalpoints, sequentialLowLevelMPS)

	dropCaches()

	then = time.Now()
	sequentialLowLevel()
	deltaccSequentialLowLevel := time.Since(then)
	ccsequentialLowLevelMPS := float64(bulkTotalpoints) / (float64(deltaccSequentialLowLevel) / 1e9) / 1e6
	fmt.Printf("COLD CACHE Sequential low level took: %s (%.3f MP/s)\n", deltaccSequentialLowLevel, ccsequentialLowLevelMPS)
	fmt.Printf(`>>> {%s, "op":"read","level":"low","pattern":"sequential","cache":"cold","time":%d,"points":%d,"mps":%f}`+"\n", cfg, deltaccSequentialLowLevel, bulkTotalpoints, ccsequentialLowLevelMPS)

	then = time.Now()
	randomLowLevel()
	deltahcRandomLowLevel := time.Since(then)
	hcRandomLowLevelMPS := float64(bulkTotalpoints) / (float64(deltahcRandomLowLevel) / 1e9) / 1e6
	fmt.Printf("HOT CACHE Random low level took: %s (%.3f MP/s)\n", deltahcRandomLowLevel, hcRandomLowLevelMPS)
	fmt.Printf(`>>> {%s, "op":"read","level":"low","pattern":"random","cache":"hot","time":%d,"points":%d,"mps":%f}`+"\n", cfg, deltahcRandomLowLevel, bulkTotalpoints, hcRandomLowLevelMPS)

	dropCaches()

	then = time.Now()
	randomLowLevel()
	deltaccRandomLowLevel := time.Since(then)
	ccRandomLowLevelMPS := float64(bulkTotalpoints) / (float64(deltaccRandomLowLevel) / 1e9) / 1e6
	fmt.Printf("COLD CACHE Random low level took: %s (%.3f MP/s)\n", deltaccRandomLowLevel, ccRandomLowLevelMPS)
	fmt.Printf(`>>> {%s, "op":"read","level":"low","pattern":"random","cache":"cold","time":%d,"points":%d,"mps":%f}`+"\n", cfg, deltaccRandomLowLevel, bulkTotalpoints, ccRandomLowLevelMPS)

	then = time.Now()
	sequentialHighLevel()
	deltahcSequentialHighLevel := time.Since(then)
	hcsequentialHighLevelMPS := float64(bulkTotalpoints) / (float64(deltahcSequentialHighLevel) / 1e9) / 1e6
	fmt.Printf("HOT CACHE Sequential high level took: %s (%.3f MP/s)\n", deltahcSequentialHighLevel, hcsequentialHighLevelMPS)
	fmt.Printf(`>>> {%s, "op":"read","level":"high","pattern":"sequential","cache":"hot","time":%d,"points":%d,"mps":%f}`+"\n", cfg, deltahcSequentialHighLevel, bulkTotalpoints, hcsequentialHighLevelMPS)

	dropCaches()
	sequentialHighLevel()
	deltaccSequentialHighLevel := time.Since(then)
	ccsequentialHighLevelMPS := float64(bulkTotalpoints) / (float64(deltaccSequentialHighLevel) / 1e9) / 1e6
	fmt.Printf("COLD CACHE Sequential high level took: %s (%.3f MP/s)\n", deltaccSequentialHighLevel, ccsequentialHighLevelMPS)
	fmt.Printf(`>>> {%s, "op":"read","level":"high","pattern":"sequential","cache":"cold","time":%d,"points":%d,"mps":%f}`+"\n", cfg, deltaccSequentialHighLevel, bulkTotalpoints, ccsequentialHighLevelMPS)

	then = time.Now()
	randomHighLevel()
	deltahcRandomHighLevel := time.Since(then)
	hcRandomHighLevelMPS := float64(bulkTotalpoints) / (float64(deltahcRandomHighLevel) / 1e9) / 1e6
	fmt.Printf("HOT CACHE Random high level took: %s (%.3f MP/s)\n", deltahcRandomHighLevel, hcRandomHighLevelMPS)
	fmt.Printf(`>>> {%s, "op":"read","level":"high","pattern":"random","cache":"hot","time":%d,"points":%d,"mps":%f}`+"\n", cfg, deltahcRandomHighLevel, bulkTotalpoints, hcRandomHighLevelMPS)

	dropCaches()

	then = time.Now()
	randomHighLevel()
	deltaccRandomHighLevel := time.Since(then)
	ccRandomHighLevelMPS := float64(bulkTotalpoints) / (float64(deltaccRandomHighLevel) / 1e9) / 1e6
	fmt.Printf("COLD CACHE Random high level took: %s (%.3f MP/s)\n", deltaccRandomHighLevel, ccRandomHighLevelMPS)
	fmt.Printf(`>>> {%s, "op":"read","level":"high","pattern":"random","cache":"cold","time":%d,"points":%d,"mps":%f}`+"\n", cfg, deltaccRandomHighLevel, bulkTotalpoints, ccRandomHighLevelMPS)

	//Test PQM insert throughput:
	//  Time 1 second inserts of 1 week of data, using 12 workers
	//Drop caches
	//Test sequential low level reads:
	//  Time raw values query for 1 month of data across 12 streams, using 12 workers
	//Drop caches
	//Test random low level reads:
	//  Using predefined seed, read out a large number of 2 minute blocks of raw data 12 streams+workers
	//Drop caches
	//Test sequential tree reads:
	//  Statistical values: 1<<37 windows of 1 month of data, 12 streams+workers sequential
	//Drop caches
	//Test random tree reads:
	//  Statistical values: 1<<37 windows of 1 month of data, 12 streams+workers random
}

func initialize() {
	btrdbconns = make([]*btrdb.BTrDB, connections)
	bulkstreams = make([]*btrdb.Stream, parallelism)
	rand.Seed(time.Now().UnixNano())
	gsetcode = rand.Int() % 0xFFFFFF
	fmt.Printf("SET CODE IS %06x\n", gsetcode)
	for i := 0; i < connections; i++ {
		db, err := btrdb.Connect(context.Background(), btrdb.EndpointsFromEnv()...)
		if err != nil {
			panic(err)
		}
		btrdbconns[i] = db
	}
}

func dropCaches() {
	ep, err := btrdbconns[0].GetAnyEndpoint(context.Background())
	if err != nil {
		panic(err)
	}
	_, err = ep.FaultInject(context.Background(), 3, nil)
	if err != nil {
		panic(err)
	}
}

func pqmInsert() {
	repctx, cancel := context.WithCancel(context.Background())
	var done int64
	go func() {
		required := parallelism * pqmHours * 60 * 60 * pqmFrequency
		for {
			time.Sleep(5 * time.Second)
			if repctx.Err() != nil {
				return
			}
			fmt.Printf("pqm insert done %d of %d (%.2f %%)\n", done, required, float64(done*100)/float64(required))
		}
	}()
	//Prepare the data sources
	datasources := make([]chan []btrdb.RawPoint, parallelism)
	for i := 0; i < parallelism; i++ {
		datasources[i] = make(chan []btrdb.RawPoint, 10)
		go func(ch chan []btrdb.RawPoint) {
			//Generate data and write to chan in batches
			start, _ := time.Parse(time.RFC3339, "2010-01-01T00:00:00+00:00")
			end := start.Add(pqmHours * time.Hour)
			cursor := start
			buf := make([]btrdb.RawPoint, 0, pqmBatchSize)
			//How much we increment
			inc := 1e9 / pqmFrequency
			for cursor.Before(end) {
				buf = append(buf, btrdb.RawPoint{Time: cursor.UnixNano(), Value: math.Sin(float64(cursor.UnixNano()) / 1e9)})
				cursor = cursor.Add(time.Duration(inc))
				if len(buf) == pqmBatchSize {
					ch <- buf
					buf = make([]btrdb.RawPoint, 0, pqmBatchSize)
				}
			}
			close(ch)
		}(datasources[i])
	}

	wg := sync.WaitGroup{}
	wg.Add(parallelism)
	for i := 0; i < parallelism; i++ {
		go func(i int) {
			db := btrdbconns[i%connections]
			ch := datasources[i]
			stream, err := db.Create(context.Background(), uuid.NewRandom(), fmt.Sprintf("alphafp/set%6x/pqm_%d", gsetcode, i), btrdb.M{"name": "stream"}, nil)
			if err != nil {
				panic(err)
			}
			for rec := range ch {
				err := stream.Insert(context.Background(), rec)
				if err != nil {
					panic(err)
				}
				atomic.AddInt64(&done, int64(len(rec)))
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	cancel()
}

func bulkInsert() {
	repctx, cancel := context.WithCancel(context.Background())
	var done int64
	go func() {
		required := parallelism * bulkDays * 24 * 60 * 60 * bulkFrequency
		for {
			time.Sleep(5 * time.Second)
			if repctx.Err() != nil {
				return
			}
			fmt.Printf("bulk insert done %d of %d (%.2f %%)\n", done, required, float64(done*100)/float64(required))
		}
	}()
	//Prepare the data sources
	datasources := make([]chan []btrdb.RawPoint, parallelism)
	for i := 0; i < parallelism; i++ {
		datasources[i] = make(chan []btrdb.RawPoint, 10)
		go func(ch chan []btrdb.RawPoint) {
			//Generate data and write to chan in batches
			start, _ := time.Parse(time.RFC3339, "2010-01-01T00:00:00+00:00")
			end := start.Add(bulkDays * 24 * time.Hour)
			cursor := start
			buf := make([]btrdb.RawPoint, 0, bulkBatchSize)
			//How much we increment
			inc := 1e9 / bulkFrequency
			for cursor.Before(end) {
				buf = append(buf, btrdb.RawPoint{Time: cursor.UnixNano(), Value: math.Sin(float64(cursor.UnixNano()) / 1e9)})
				cursor = cursor.Add(time.Duration(inc))
				if len(buf) == bulkBatchSize {
					ch <- buf
					buf = make([]btrdb.RawPoint, 0, bulkBatchSize)
				}
			}
			close(ch)
		}(datasources[i])
	}

	wg := sync.WaitGroup{}
	wg.Add(parallelism)
	for i := 0; i < parallelism; i++ {
		go func(i int) {
			db := btrdbconns[i%connections]
			ch := datasources[i]
			stream, err := db.Create(context.Background(), uuid.NewRandom(), fmt.Sprintf("alphafp/set%6x/bulk_%d", gsetcode, i), btrdb.M{"name": "stream"}, nil)
			if err != nil {
				panic(err)
			}
			bulkstreams[i] = stream
			for rec := range ch {
				err := stream.Insert(context.Background(), rec)
				if err != nil {
					panic(err)
				}
				atomic.AddInt64(&done, int64(len(rec)))
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	cancel()
}

func sequentialLowLevel() {
	repctx, cancel := context.WithCancel(context.Background())
	var done int64
	required := parallelism * bulkDays * 24 * 60 * 60 * bulkFrequency
	go func() {

		for {
			time.Sleep(5 * time.Second)
			if repctx.Err() != nil {
				return
			}
			fmt.Printf("sequential low level read done %d of %d (%.2f %%)\n", done, required, float64(done*100)/float64(required))
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(parallelism)
	for i := 0; i < parallelism; i++ {
		go func(i int) {
			stream := bulkstreams[i]
			start, _ := time.Parse(time.RFC3339, "2010-01-01T00:00:00+00:00")
			end := start.Add(bulkDays * 24 * time.Hour)
			points, _, err := stream.RawValues(context.Background(), start.UnixNano(), end.UnixNano(), btrdb.LatestVersion)
			for _ = range points {
				atomic.AddInt64(&done, 1)
			}
			if e := <-err; e != nil {
				panic(e)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	fmt.Printf("sequential low level read done %d of %d (%.2f %%)\n", done, required, float64(done*100)/float64(required))
	cancel()
}

func sequentialHighLevel() {
	repctx, cancel := context.WithCancel(context.Background())
	var done int64
	required := parallelism * bulkDays * 24 * 60 * 60 * bulkFrequency
	go func() {

		for {
			time.Sleep(5 * time.Second)
			if repctx.Err() != nil {
				return
			}
			fmt.Printf("sequential low level read done %d of %d (%.2f %%)\n", done, required, float64(done*100)/float64(required))
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(parallelism)
	for i := 0; i < parallelism; i++ {
		go func(i int) {
			stream := bulkstreams[i]
			start, _ := time.Parse(time.RFC3339, "2010-01-01T00:00:00+00:00")
			end := start.Add(bulkDays * 24 * time.Hour)
			points, _, err := stream.AlignedWindows(context.Background(), start.UnixNano(), end.UnixNano()+1<<37, 37, btrdb.LatestVersion)
			for p := range points {
				atomic.AddInt64(&done, int64(p.Count))
			}
			if e := <-err; e != nil {
				panic(e)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	fmt.Printf("sequential low level read done %d of %d (%.2f %%)\n", done, required, float64(done*100)/float64(required))
	cancel()
}

func randomLowLevel() {

	periodNanos := bulkDays * 24 * 60 * 60 * 1000000000
	windows := periodNanos / 60e9
	src := rand.NewSource(39)
	rnd := rand.New(src)
	windowOrderings := make([][]int, parallelism)
	for i := 0; i < parallelism; i++ {
		windowOrderings[i] = rnd.Perm(windows)
	}

	repctx, cancel := context.WithCancel(context.Background())
	var done int64
	required := parallelism * bulkDays * 24 * 60 * 60 * bulkFrequency
	go func() {
		for {
			time.Sleep(5 * time.Second)
			if repctx.Err() != nil {
				return
			}
			fmt.Printf("random low level read done %d of %d (%.2f %%)\n", done, required, float64(done*100)/float64(required))
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(parallelism)
	for i := 0; i < parallelism; i++ {
		go func(i int) {
			stream := bulkstreams[i]
			start, _ := time.Parse(time.RFC3339, "2010-01-01T00:00:00+00:00")
			windows := windowOrderings[i]
			for _, w := range windows {
				rangestart := start.Add(time.Duration(w * 60e9))
				rangeend := rangestart.Add(time.Duration(60e9))
				points, _, err := stream.RawValues(context.Background(), rangestart.UnixNano(), rangeend.UnixNano(), btrdb.LatestVersion)
				for _ = range points {
					atomic.AddInt64(&done, 1)
				}
				if e := <-err; e != nil {
					panic(e)
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	fmt.Printf("random low level read done %d of %d (%.2f %%)\n", done, required, float64(done*100)/float64(required))
	cancel()
}

func randomHighLevel() {

	periodNanos := bulkDays * 24 * 60 * 60 * 1000000000
	windows := periodNanos >> 37
	src := rand.NewSource(37)
	rnd := rand.New(src)
	windowOrderings := make([][]int, parallelism)
	for i := 0; i < parallelism; i++ {
		windowOrderings[i] = rnd.Perm(windows + 1)
	}

	repctx, cancel := context.WithCancel(context.Background())
	var done int64
	required := parallelism * bulkDays * 24 * 60 * 60 * bulkFrequency
	go func() {
		for {
			time.Sleep(5 * time.Second)
			if repctx.Err() != nil {
				return
			}
			fmt.Printf("random high level read done %d of %d (%.2f %%)\n", done, required, float64(done*100)/float64(required))
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(parallelism)
	for i := 0; i < parallelism; i++ {
		go func(i int) {
			stream := bulkstreams[i]
			start, _ := time.Parse(time.RFC3339, "2010-01-01T00:00:00+00:00")
			windows := windowOrderings[i]
			for _, w := range windows {
				rangestart := start.Add(time.Duration(w * (1 << 37)))
				rangeend := rangestart.Add(time.Duration(1 << 37))
				points, _, err := stream.AlignedWindows(context.Background(), rangestart.UnixNano(), rangeend.UnixNano(), 37, btrdb.LatestVersion)
				for p := range points {
					atomic.AddInt64(&done, int64(p.Count))
				}
				if e := <-err; e != nil {
					panic(e)
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	fmt.Printf("random high level read done %d of %d (%.2f %%)\n", done, required, float64(done*100)/float64(required))
	cancel()
}
