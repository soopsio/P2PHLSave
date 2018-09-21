package main

import "crypto/md5"
import "flag"
import "fmt"
import "time"
import "net/http"
import "encoding/json"
import "log"
import "github.com/gorilla/mux"
import "github.com/go-redis/redis"

/*
tracker

https://api.cdnbye.com/v1/channel
POST
Content-Type: text/plain;charset=UTF-8
Origin: null
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.92 Safari/537.36
{"device":"PC","netType":"","version":"0.2.5","tag":"0.10.0","channel":"JTJGMjAxODA5MTYlMkYyMzE3XzQyYWU3YjUxJTJGaW5kZXglNUJ2MSU1RA==","ts":1537328692}


curl -v -X POST localhost:8787/channel -H "content-type:application/json" -d '{"device":"PC","netType":"","version":"0.2.5","tag":"0.10.0","channel":"JTJGMjAxODA5MTYlMkYyMzE3XzQyYWU3YjUxJTJGaW5kZXglNUJ2MSU1RA==","ts":1537328692}'

{
  "ret": 0,
  "name": "channel",
  "data": {
    "id": "3e8b185c7075bbef",
    "v": "b0a44ff5",
    "report_interval": 15,
    "peers": [
      {
        "id": "11cd6ebd459b9cde"
      }
    ]
  }
}


POST
curl -v -X POST  -H "content-type:application/json" http://localhost:8787/channel/JTJGMjAxODA5MTYlMkYyMzE3XzQyYWU3YjUxJTJGaW5kZXglNUJ2MSU1RA==/node/3e8b185c7075bbef/peers -d ''

{
  "ret": 0,
  "name": "peers",
  "data": {
    "peers": [
      {
        "id": "11cd6ebd459b9cde"
      }
    ]
  }
}

POST

curl -v -X POST  -H "content-type:application/json" http://localhost:8787/channel/JTJGMjAxODA5MTYlMkYyMzE3XzQyYWU3YjUxJTJGaW5kZXglNUJ2MSU1RA==/node/3e8b185c7075bbef/stats -d '{"p2p":357}'


https://api.cdnbye.com/v1/channel/JTJGMjAxODA5MTYlMkYyMzE3XzQyYWU3YjUxJTJGaW5kZXglNUJ2MSU1RA==/node/3e8b185c7075bbef/stats
{"p2p":357}

{"http":693}


{
  "ret": 0,
  "name": "stats",
  "data": null
}

*/
var addr = flag.String("addr", ":8787", "http service address")
var redis_host = flag.String("redis_host", "localhost:6379", "redis host to connect")
var peer_ttl = flag.Int64("peer_ttl", 300, "peer live timeout")

type Announce struct {
	channel string
	device  string
	netType string
	tag     string
	version string
	ts      uint64
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func generate_peer_id(announce Announce) string {
	s := fmt.Sprintf("%s:%s:%s", announce.channel, announce.ts, time.Now().Unix())
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))[0:16]
}
func register(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	// fmt.Println(r.URL.Path)
	// fmt.Println(r.Header["Content-Type"])
	if r.ContentLength > 0 {
		var body Announce
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		// fmt.Println(body)
		channel := body.channel
		peer_id := generate_peer_id(body)
		update_peer(channel, peer_id)
		ret := 0
		fmt.Printf("channel user added %s \n", channel)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ret":  ret,
			"name": "channel",
			"data": map[string]interface{}{
				"id":              peer_id,
				"v":               "verify_code", // @todo
				"report_interval": 15,
				"peers":           get_peers(channel, peer_id), //@todo client should follow this change
			},
		})
	} else {
		fmt.Fprint(w, "POST with body please")
	}
}
func stats(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	channel := params["channel"]
	peer_id := params["peer"]
	// w.Write([]byte("channel " + channel + "  peer " + peer))
	fmt.Printf("channel %s  peer %s stats \n", channel, peer_id)
	update_peer(channel, peer_id)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enableCors(&w)
	// w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ret":  0,
		"name": "stats",
		"data": nil,
	})
}
func peers(w http.ResponseWriter, r *http.Request) {
	// redis_client := redis.NewClient(&redis.Options{
	// 	Addr:     "localhost:6379",
	// 	Password: "", // no password set
	// 	DB:       0,  // use default DB
	// })
	params := mux.Vars(r)
	channel := params["channel"]
	peer_id := params["peer"]
	// w.Write([]byte("channel " + channel + "  peer " + peer))
	fmt.Printf("channel %s  peer %s peers \n", channel, peer_id)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enableCors(&w)
	// w.WriteHeader(http.StatusCreated)
	// peers, _ := redis_client.LRange("peers", 0, 30).Result()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ret":  0,
		"name": "peers",
		"data": map[string]interface{}{
			"peers": get_peers(channel, peer_id), //@todo client should follow this change
		},
	})
}
func echo(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	fmt.Println(r.URL.Path)
	fmt.Println(r.Header["Content-Type"])
	if r.ContentLength > 0 {
		var body map[string]*json.RawMessage
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		json.NewEncoder(w).Encode(body)
	} else {
		fmt.Fprint(w, "get request")
	}
}
func home(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Tracker only accept api request",
		"ret":     0,
	})
}

func update_peer(channel string, peer_id string) {
	ts := time.Now().Unix()
	redis_client := redis.NewClient(&redis.Options{
		Addr:     *redis_host,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	redis_client.ZAdd("peers:"+channel, redis.Z{Score: float64(0 - ts), Member: peer_id})
}
func get_peers(channel string, peer_id string) []string {
	ts := time.Now().Unix()
	redis_client := redis.NewClient(&redis.Options{
		Addr:     *redis_host,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	keys, err := redis_client.ZRangeByScore("peers:"+channel, redis.ZRangeBy{
		Min:    fmt.Sprintf("%f", float64(0-ts)-float64(*peer_ttl)),
		Max:    "+inf", //fmt.Sprintf("%f", float64(0-ts)),
		Offset: 0,
		Count:  5}).Result()
	// pong, err := redis_client.Ping().Result()
	if err != nil {
		fmt.Printf("ping has response %s", keys)
	}
	peers := keys[:0]
	for _, x := range keys {
		// fmt.Println("key", x, peer_id, x == peer_id)
		if x != peer_id {
			peers = append(peers, x)
		}
	}
	return peers
}
func redis_test(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	// redis_client := redis.NewClient(&redis.Options{
	// 	Addr:     "localhost:6379",
	// 	Password: "", // no password set
	// 	DB:       0,  // use default DB
	// })
	params := mux.Vars(r)
	// ts := time.Now().Unix()
	channel := "test"
	peer_id := params["key"]
	update_peer(channel, peer_id)
	keys := get_peers(channel, peer_id)
	ret := 0
	// if err != nil {
	// 	fmt.Printf("ping has response %s", keys)
	// 	ret = 1
	// }
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": keys,
		"ret":     ret,
	})
}
func main() {
	flag.Parse()
	channel_router := mux.NewRouter()
	channel_router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	channel_router.HandleFunc("/", home)
	channel_router.HandleFunc("/redis/{key}", redis_test)
	channel_router.HandleFunc("/channel", register).Methods("POST")
	channel_router.HandleFunc("/channel/{channel}/node/{peer}/stats", stats).Methods("POST")
	channel_router.HandleFunc("/channel/{channel}/node/{peer}/peers", peers).Methods("POST")
	channel_router.HandleFunc("/echo", echo)
	http.Handle("/", channel_router)
	fmt.Sprintf("curl localhost%s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
