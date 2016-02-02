package main

// To run:
// go build
// go test -v

import (
	`bytes`
	`encoding/json`
	`fmt`
	`io/ioutil`
	`math/rand`
	`net/http`
	`net/url`
	`strconv`
	`testing`
	`strings`
	`time`
	`os`
)

const (
    SocketTimeoutMs = 5000
)

type CachePair struct {
	Key   interface{} `json:"key"`
	Value interface{} `json:"value"`
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func getServiceTestAddress() string {
	host := strings.TrimSpace(os.Getenv("DEVTEST_HOST"))
	if host == "" {
		host = "localhost"
	}

	port := strings.TrimSpace(os.Getenv("DEVTEST_PORT"))
	if port == "" {
		port = "8088"
	}

	res := "http://"+string(host)+":"+string(port)+"/cache/"

	return res
}

func TestChallenge1Post(t *testing.T) {
	// Start from a clean slate.
	deleteAll(t)

	// Populate a single key / value pair via a POST call.
	cp1 := &CachePair{Key: "foo", Value: "bar"}

	// Status code 201 returned with a link to new cache item's key.
	post(t, cp1)
}

func TestChallenge1GetAll(t *testing.T) {
	// Start from a clean slate.
	deleteAll(t)

	// Populate multiple cache items via post calls.
	cp1 := &CachePair{Key: "alpha", Value: "beta"}
	post(t, cp1)

	cp2 := &CachePair{Key: 100, Value: "dolla billllllllz!"}
	post(t, cp2)

	cp3 := &CachePair{Key: 100.001, Value: true}
	post(t, cp3)

	// Get cache contents, compare to our in memory verison.
	cpairs := []*CachePair{cp1, cp2, cp3}
	getAll(t, cpairs)
}

func TestChallenge1DeleteAll(t *testing.T) {
	// Start from a clean slate.
	deleteAll(t)

	// Populate multiple cache items via post calls.
	cp1 := &CachePair{Key: "arugula", Value: false}
	post(t, cp1)

	cp2 := &CachePair{Key: -1, Value: -0.0005}
	post(t, cp2)

	// Delete everything. Dieeee, cache scum!
	deleteAll(t)

	// Cache should be empty now.
	getAll(t, []*CachePair{})
}

func TestChallenge1GetKey(t *testing.T) {
	// Populate a single cache item.
	cp1 := &CachePair{Key: "bazooty", Value: "tootyfruity"}
	post(t, cp1)

	// Get the key we just populated.
	getKey(t, cp1)
}

func TestChallenge1Put(t *testing.T) {
	// Populate a single cache item.
	cp1 := &CachePair{Key: 99, Value: 100}
	post(t, cp1)

	// Update the value via a put call.
	cp1.Value = 101
	putKey(t, cp1)

	// Get the key we just populated.
	getKey(t, cp1)
}

func TestChallenge1DeleteKey(t *testing.T) {
	// Start from a clean slate.
	deleteAll(t)

	// Populate a couple items via post.
	cp1 := &CachePair{Key: true, Value: "rainier"}
	post(t, cp1)

	cp2 := &CachePair{Key: "haggis", Value: 9999999.999}
	post(t, cp2)

	// Delete just that key. Other stuff should live.
	deleteKey(t, cp1)

	// Cache should be empty now.
	getAll(t, []*CachePair{cp2})
}

func TestChallenge1PostUnhappy(t *testing.T) {
	// Start from a clean slate.
	deleteAll(t)

	// Populate a couple items via post.
	cp1 := &CachePair{Key: "ayyyyyeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee!!!", Value: "look at that long key"}
	post(t, cp1)
	postForStatus(t, cp1, http.StatusConflict)
}

func TestChallenge1GetUnhappy(t *testing.T) {
	// Start from a clean slate.
	deleteAll(t)

	// We should not find this key.
	cp_rando := &CachePair{Key: "Rando", Value: "Calrissian"}
	getKeyForStatus(t, cp_rando, http.StatusNotFound)
}

func TestChallenge1PutUnhappy(t *testing.T) {
	// Clean slate.
	deleteAll(t)

	// Attempt to update a non-existent key.
	cp_unknown := &CachePair{Key: "false", Value: "barf"}
	putKeyForStatus(t, cp_unknown, http.StatusNotFound)
}

func TestChallenge1DeleteUnhappy(t *testing.T) {
	// Clean slate.
	deleteAll(t)

	// Attempt to delete a non-existent key.
	cp_unknown := &CachePair{Key: "?", Value: false}
	deleteKeyForStatus(t, cp_unknown, http.StatusNotFound)
}

func TestChallenge1CrazyLength(t *testing.T) {
	deleteAll(t)

	// Populate a single large cache item.
	longKey := randomString(10000)
	longValue := randomString(10000)
	cp1 := &CachePair{Key: longKey, Value: longValue}
	post(t, cp1)

	// Get the key we just populated.
	getKey(t, cp1)
}

func TestChallenge1CrazyCharset(t *testing.T) {
	deleteAll(t)

	// Let's get Unicodical!
	cp1 := &CachePair{Key: "世ƁǆΏ", Value: "界ĘĶƁΔӜӦ‡"}
	post(t, cp1)

	// Did that work?
	getKey(t, cp1)
}

func TestChallenge1CrazyTypes(t *testing.T) {
	deleteAll(t)

	// Populate multiple cache items via post calls. Sooo similar.
	cp1 := &CachePair{Key: "123", Value: "one twenty three"}
	post(t, cp1)

	cp2 := &CachePair{Key: 123, Value: 123}
	post(t, cp2)

	cp3 := &CachePair{Key: 123.0000001, Value: 123.0000000}
	post(t, cp3)

	cp4 := &CachePair{Key: 122.9999999999999, Value: 122.9999999999}
	post(t, cp4)

	// Get cache contents, compare to our in memory verison.
	cpairs := []*CachePair{cp1, cp2, cp3, cp4}
	getAll(t, cpairs)
}

func TestChallenge1CrazyJson(t *testing.T) {
	deleteAll(t)

	// We will throw some fun, JSON-y strings into the cache.
	cp1 := &CachePair{Key: "{\"holiday\":\"thanksgiving\"}", Value: "gobble gobble!"}
	post(t, cp1)

	getAll(t, []*CachePair{cp1})
}

func TestChallenge1CrazySize(t *testing.T) {
	deleteAll(t)

	// Let's make a big cache with big strings. Does the cache hold up?
	const biggun_size = 1000
	var biggun_cache [biggun_size]*CachePair
	for i := 0; i < biggun_size; i++ {
		cp := &CachePair{Key: randomString(biggun_size), Value: randomString(biggun_size)}
		biggun_cache[i] = cp
		post(t, cp)
	}

	getAll(t, biggun_cache[0:biggun_size])
}

func TestChallenge2Synchronous(t *testing.T) {
	deleteAll(t)

	const loop_size = 100
	cp := &CachePair{Key: randomString(loop_size), Value: randomString(loop_size)}
	post(t, cp)

	// We make 100 calls to the cache to get this key. Once we're through this loop
	// it should be dropped from the cache as if it were hot.
	for i := 0; i < loop_size; i++ {
		getKey(t, cp)
	}

	// Cache is empty, yes?
	getAll(t, []*CachePair{})
}

func TestChallenge2Asynchronous(t *testing.T) {
	deleteAll(t)

	const loop_size = 100
	cp := &CachePair{Key: randomString(loop_size), Value: randomString(loop_size)}
	post(t, cp)

	sent := make(chan bool)

	// We make 100 concurrent calls to get this key.
	for i := 0; i < loop_size; i++ {
		go getKeyAsync(t, cp, sent)
	}

	// Wait until all 100 workers have sent their request.
	for i := 0; i < loop_size; i++ {
		<-sent
	}

	getAll(t, []*CachePair{})
}

func TestChallenge2Crazy(t *testing.T) {
	deleteAll(t)

	const loop_size = 50
	cp := &CachePair{Key: randomString(loop_size), Value: randomString(loop_size)}
	post(t, cp)

	sent := make(chan bool)

	// We make 100 concurrent calls to get this key, spread across different get calls.
	for i := 0; i < loop_size; i++ {
		go getKeyAsync(t, cp, sent)
		go getAllAsync(t, []*CachePair{cp}, sent)
	}

	// Wait until all 100 workers have sent their request.
	for i := 0; i < loop_size * 2; i++ {
		<-sent
	}

	getAll(t, []*CachePair{})
}

// this test is to test persistance not ordering of scheduling requests
func TestChallenge3(t *testing.T) {
	deleteAll(t)

	const cache_size = 1000
	const calls_per_key = 10
	const halt_seconds = 10
	const sleep_millis = 100

	// Create 1000 keys.
	var cache [cache_size]*CachePair
	for i := 0; i < cache_size; i++ {
		cp := &CachePair{Key: randomString(10)+fmt.Sprintf("%d", i), Value: randomFloat()}
		cache[i] = cp
		post(t, cp)
	}
    fmt.Printf("Cache keys created.\n")

	// For each key, make 20 puts with teeny tiny changes.
	sent := make(chan bool)
	for j := 0; j < calls_per_key; j++ {
		for i := 0; i < cache_size; i++ {		
			cache[i].Value = cache[i].Value.(float64) + 1
			go putKeyAsync(t, cache[i], sent)
		}

		// Allow all calls to finish. We're accounting for 20 puts for each key.
		for i := 0; i < cache_size; i++ {
			<-sent
		}

		// Sleep for 10 milliseconds, just to give the service a chance.
		// time.Sleep(time.Millisecond * sleep_millis)
	}
    fmt.Printf("Cache keys updated.\n")

	// Halt execution for 30 seconds to allow for server bounce.
	fmt.Printf("\nChallenge 3 test case halted for %d seconds! Quick, bounce the cache service!\n", halt_seconds)
	time.Sleep(time.Second * halt_seconds)

	// Make a get call and compare in-memory vs. service cache.
	getAll(t, cache[0:cache_size])
}

func TestChallenge4(t *testing.T) {
    deleteAll(t)

    const secs_to_run = 60
    const sla_millis = 100
    const desired_sla_perc = .05
	const sleep_millis = 100
    const cache_size = 100
    const reqs_per_sec = 100
    const expected_reqs = (secs_to_run * reqs_per_sec)
	
    var cache [cache_size]*CachePair
    cache_index := 0
    total_reqs := 0

	response := make(chan int64)

    // Iterate through all of our responses. What happened?
    sla_violations := 0

    for true == true {
        // Every 1000 iterations, get the cache, blow it away, start over.
        if total_reqs  > 0 && total_reqs % cache_size == 0 {
			// Allow all calls to finish. We're accounting for 20 puts for each key.
			for i := 0; i < cache_index * 2; i++ {
				resp_time := <-response
		        if resp_time > sla_millis {
		            sla_violations++
		        }
			}

            // Sleep briefly, just to give the cache a chance.
            time.Sleep(time.Millisecond * sleep_millis)
            fmt.Printf("Request block %d.\n", total_reqs)

            // Synchronous call to getAll.
            //go getAllAsyncForResponseTime(t, cache[0:cache_index], response)
            getAll(t, cache[0:cache_index])
            
            // Synchronous call to deleteAll.
            //go deleteAllAsyncForResponseTime(t, response)
            deleteAll(t)
            
            cache_index = 0
            total_reqs += 2
        }

        // Create an item in the cache, update it right away.
		cp := &CachePair{Key: randomString(cache_size), Value: randomString(cache_size)}
		cache[cache_index] = cp
        go postPutAsyncForResponseTime(t, cp, response)
        cache_index++
        total_reqs += 2

        if total_reqs == expected_reqs {
            break
        }
        
        time.Sleep(time.Millisecond * (1000 / reqs_per_sec))
    }

    time.Sleep(time.Millisecond * SocketTimeoutMs)

    // hackhackhack
    for i := 0; i < cache_index * 2; i++ {
        resp_time := <-response
        if resp_time > sla_millis {
            sla_violations++
        }
	}

    fmt.Printf("Violations is %d, total reqs us %d", sla_violations, total_reqs)
    actual_sla_perc := float64(sla_violations) / float64(total_reqs)
    fmt.Printf("\nYou met the SLA for %.2f percent of requests.\n", (1 - actual_sla_perc) * 100)
    if (actual_sla_perc >= desired_sla_perc) {
        t.Errorf("Test failed! You violated the SLA %.2f of the time, where we required %.2f.", actual_sla_perc, desired_sla_perc)
    }
}

func deleteAllAsyncForResponseTime(t *testing.T, response chan<- int64) {
    startMillis := time.Now().UnixNano() / int64(time.Millisecond)
    deleteAll(t)
    endMillis := time.Now().UnixNano() / int64(time.Millisecond)
    response <- (endMillis - startMillis)
}

func deleteAll(t *testing.T) {
	expectedStatus := http.StatusNoContent
	req, err := http.NewRequest("DELETE", getServiceTestAddress(), nil)
    req.Close = true
	if err != nil {
		t.Errorf("Delete call failed: %s", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("Error reading response body: %s %s", err, getServiceTestAddress())
	} else {
		resp.Body.Close()
	}

	if resp.StatusCode != expectedStatus {
		t.Errorf("Response code of %d doesn't match expected %d.", resp.StatusCode, expectedStatus)
	}
}

func deleteKey(t *testing.T, cp *CachePair) {
	deleteKeyForStatus(t, cp, http.StatusNoContent)
}

func deleteKeyForStatus(t *testing.T, cp *CachePair, expectedStatus int) {
	endpoint := fmt.Sprintf("%s%s", getServiceTestAddress(), getUrlFriendlyKey(cp))
	req, err := http.NewRequest("DELETE", endpoint, nil)
    req.Close = true
	if err != nil {
		t.Errorf("Delete call failed: %s", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("Error reading response body: %s", err)
	} else {
		resp.Body.Close()
	}

	if resp.StatusCode != expectedStatus {
		t.Errorf("Response code of %d doesn't match expected %d.", resp.StatusCode, expectedStatus)
	}
}

func post(t *testing.T, cp *CachePair) {
	postForStatus(t, cp, http.StatusCreated)
}

func postAsyncForResponseTime(t *testing.T, cp *CachePair, response chan<- int64) {
    startMillis := time.Now().UnixNano() / int64(time.Millisecond)
    post(t, cp)
    endMillis := time.Now().UnixNano() / int64(time.Millisecond)
    response <- (endMillis - startMillis)
}

func postPutAsyncForResponseTime(t *testing.T, cp *CachePair, response chan<- int64) {
    // Create a cache dealie.
    startMillis := time.Now().UnixNano() / int64(time.Millisecond)
    post(t, cp)
    endMillis := time.Now().UnixNano() / int64(time.Millisecond)
    response <- (endMillis - startMillis)

    // Update said cache dealie.
    cp.Value = randomString(1000)
    startMillis = time.Now().UnixNano() / int64(time.Millisecond)
    putKey(t, cp)
    endMillis = time.Now().UnixNano() / int64(time.Millisecond)
    response <- (endMillis - startMillis)
}

func postForStatus(t *testing.T, cp *CachePair, expectedStatus int) {
	cpJson, err := json.Marshal(cp)
	if err != nil {
		t.Errorf("Unable to marshal %s into json.", cp)
	}

	resp, err := http.Post(getServiceTestAddress(), "application/json", bytes.NewReader(cpJson))
	if err != nil {
		t.Errorf("Initial connection failed: %s", err)
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading response body: %s", err)
	}

	if resp.StatusCode != expectedStatus {
		t.Errorf("Response code of %d doesn't match expected %d.", resp.StatusCode, expectedStatus)
	}
}

func putKey(t *testing.T, cp *CachePair) {
	putKeyForStatus(t, cp, http.StatusNoContent)
}

func putKeyAsync(t *testing.T, cp *CachePair, sent chan<- bool) {
	putKey(t, cp)
	sent <- true
}

func putKeyAsyncForResponseTime(t *testing.T, cp *CachePair, response chan<- int64) {
    startMillis := time.Now().UnixNano() / int64(time.Millisecond)
    putKey(t, cp)
    endMillis := time.Now().UnixNano() / int64(time.Millisecond)
    response <- (endMillis - startMillis)
}

func putKeyForStatus(t *testing.T, cp *CachePair, expectedStatus int) {
	endpoint := fmt.Sprintf("%s%s", getServiceTestAddress(), getUrlFriendlyKey(cp))
	cpJson, err := json.Marshal(cp)

	req, err := http.NewRequest("PUT", endpoint, bytes.NewReader(cpJson))
	if err != nil {
		t.Errorf("Put call failed: %s", err)
	}
    req.Close = true

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("Error reading response body: %s", err)
	} else {
		resp.Body.Close()
	}

	if resp != nil && resp.StatusCode != expectedStatus {
		t.Errorf("Response code of %d doesn't match expected %d.", resp.StatusCode, expectedStatus)
	}
}

func getKey(t *testing.T, cp *CachePair) {
	getKeyForStatus(t, cp, http.StatusOK)
}

func getKeyAsync(t *testing.T, cp *CachePair, sent chan<- bool) {
	getKeyForStatus(t, cp, http.StatusOK)
	sent <- true
}

func getKeyForStatus(t *testing.T, cp *CachePair, expectedStatus int) {
	endpoint := fmt.Sprintf("%s%s", getServiceTestAddress(), getUrlFriendlyKey(cp))
	resp, err := http.Get(endpoint)

	if err != nil {
		t.Errorf("Initial connection failed: %s", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		t.Errorf("Response code of %d doesn't match expected %d.", resp.StatusCode, expectedStatus)
		return
	}

	// If we're testing the unhappy path, bail out before all of the cache comparison.
	if expectedStatus != http.StatusOK {
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading response body: %s", err)
	}

	var data CachePair
	err = json.Unmarshal(body, &data)

	if err != nil {
		// If we can't unmarshal into a single pair, we try to unmarshal into an array
		// of pairs. Driven by some uncertainty in requirements vs. Kirkwood's implementation.
		var cacheArr []CachePair
		arrErr := json.Unmarshal(body, &cacheArr)
		if arrErr != nil {
			t.Errorf("Can't unmarshal into a single pair or an array of pairs. Giving up! Single pair unmarshal error: %s; array of pairs unmarshal error: %s.", err, arrErr)
		} else {
			data = cacheArr[0]
		}
	}

	compareCaches(t, []*CachePair{cp}, []CachePair{data})
}

func getAllAsyncForResponseTime(t *testing.T, cpairs []*CachePair, response chan<- int64) {
    startMillis := time.Now().UnixNano() / int64(time.Millisecond)
    getAll(t, cpairs)
    endMillis := time.Now().UnixNano() / int64(time.Millisecond)
    response <- (endMillis - startMillis)
}

func getAllAsync(t *testing.T, cpairs []*CachePair, sent chan<- bool) {
	getAll(t, cpairs)
	sent <- true
}

func getAll(t *testing.T, cpairs []*CachePair) {
	expectedStatus := http.StatusOK
	resp, err := http.Get(getServiceTestAddress())

	if err != nil {
		t.Errorf("Initial connection failed: %s", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		t.Errorf("Response code of %d doesn't match expected %d.", resp.StatusCode, expectedStatus)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading response body: %s", err)
	}

	var cacheMap map[string][]CachePair
	err = json.Unmarshal(body, &cacheMap)

	if err != nil {
		t.Errorf("Can't unmarshal response into a map of string -> cache pair: %s.", err)
	}

	if _, ok := cacheMap["cache"]; !ok {
		t.Errorf("Could not find cache key in response. Response equals %s.\n", cacheMap)
	}

	compareCaches(t, cpairs, cacheMap["cache"])
}

func compareCaches(t *testing.T, memoryPairs []*CachePair, serverPairs []CachePair) {
	// Put both the in-memory cache and server cache into key -> value format for
	// soopa fast indexing.
	mem_cache := make(map[interface{}]interface{})
	for _, cp := range memoryPairs {
		k := cp.Key
		v := cp.Value

		// int's get munged into float's as part of the marshal / unmarshal dance.
		if k_int, ok := cp.Key.(int); ok {
			k = float64(k_int)
		}

		if v_int, ok := cp.Value.(int); ok {
			v = float64(v_int)
		}

		mem_cache[k] = v
	}

	// No type magic for server cache, let's just take what was unmarshaled.
	serv_cache := make(map[interface{}]interface{})
	for _, cp := range serverPairs {
		serv_cache[cp.Key] = cp.Value
	}

	if len(mem_cache) != len(serv_cache) {
		t.Errorf("In memory cache and server cache have different lengths! Actual: %d, expected %d.", len(serv_cache), len(mem_cache))
	}

	for k, _ := range mem_cache {
		mem_v := mem_cache[k]
		if serv_v, ok := serv_cache[k]; ok {
			if mem_v != serv_v {
				t.Errorf("Values for key %s differ across in-memory and server caches. Actual: %s, expected: %s.", k, serv_cache, mem_cache)
			}
		} else {
			t.Errorf("Key %s found in memory cache, but not in server cache.", k)
		}
	}
}

func getUrlFriendlyKey(cp *CachePair) string {
	if k_int, ok := cp.Key.(int); ok {
		return strconv.Itoa(k_int)
	}
	if k_float, ok := cp.Key.(float64); ok {
		return strconv.FormatFloat(k_float, 'f', 6, 64)
	}
	if k_bool, ok := cp.Key.(bool); ok {
		if k_bool {
			return "true"
		} else {
			return "false"
		}
	}
	return url.QueryEscape(cp.Key.(string))
}

func randomString(n int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func randomFloat() float64 {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Float64()
}
