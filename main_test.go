package main

import (
	`bytes`
	`encoding/json`
	`fmt`
	`io/ioutil`
	`net/http`
	`strconv`
	`testing`
)

const (
	LocalHost = "http://localhost:8088/cache/"
)

type CachePair struct {
	Key   interface{} `json:"key"`
	Value interface{} `json:"value"`
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

func deleteAll(t *testing.T) {
	expectedStatus := http.StatusNoContent
	req, err := http.NewRequest("DELETE", LocalHost, nil)
	if err != nil {
		t.Errorf("Delete call failed: %s", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("Error reading response body: %s", err)
	}

	if resp.StatusCode != expectedStatus {
		t.Errorf("Response code of %s doesn't match expected %s.", resp.StatusCode, expectedStatus)
	}
}

func deleteKey(t *testing.T, cp *CachePair) {
	expectedStatus := http.StatusNoContent
	endpoint := fmt.Sprintf("%s%s", LocalHost, getUrlFriendlyKey(cp))
	req, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		t.Errorf("Delete call failed: %s", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("Error reading response body: %s", err)
	}

	if resp.StatusCode != expectedStatus {
		t.Errorf("Response code of %s doesn't match expected %s.", resp.StatusCode, expectedStatus)
	}
}

func post(t *testing.T, cp *CachePair) {
	expectedStatus := http.StatusCreated
	cpJson, err := json.Marshal(cp)
	if err != nil {
		t.Errorf("Unable to marshal %s into json.", cp)
	}

	resp, err := http.Post(LocalHost, "application/json", bytes.NewReader(cpJson))
	if err != nil {
		t.Errorf("Initial connection failed: %s", err)
	}

	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading response body: %s", err)
	}

	if resp.StatusCode != expectedStatus {
		t.Errorf("Response code of %s doesn't match expected %s.", resp.StatusCode, expectedStatus)
	}
}

func putKey(t *testing.T, cp *CachePair) {
	expectedStatus := http.StatusNoContent
	endpoint := fmt.Sprintf("%s%s", LocalHost, getUrlFriendlyKey(cp))
	cpJson, err := json.Marshal(cp)

	req, err := http.NewRequest("PUT", endpoint, bytes.NewReader(cpJson))
	if err != nil {
		t.Errorf("Put call failed: %s", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("Error reading response body: %s", err)
	}

	if resp.StatusCode != expectedStatus {
		t.Errorf("Response code of %s doesn't match expected %s.", resp.StatusCode, expectedStatus)
	}
}

func getKey(t *testing.T, cp *CachePair) {
	expectedStatus := http.StatusOK
	endpoint := fmt.Sprintf("%s%s", LocalHost, getUrlFriendlyKey(cp))
	resp, err := http.Get(endpoint)

	if err != nil {
		t.Errorf("Initial connection failed: %s", err)
	}

	if resp.StatusCode != expectedStatus {
		t.Errorf("Response code of %s doesn't match expected %s.", resp.StatusCode, expectedStatus)
	}

	defer resp.Body.Close()
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

func getAll(t *testing.T, cpairs []*CachePair) {
	expectedStatus := http.StatusOK
	resp, err := http.Get(LocalHost)

	if err != nil {
		t.Errorf("Initial connection failed: %s", err)
	}

	if resp.StatusCode != expectedStatus {
		t.Errorf("Response code of %s doesn't match expected %s.", resp.StatusCode, expectedStatus)
	}

	defer resp.Body.Close()
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
	return fmt.Sprintf("%s", cp.Key)
}
