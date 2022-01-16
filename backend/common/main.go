package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	fmt.Println("11111111111111111")

	maps := make([]map[string]interface{}, 6)
	map1 := make(map[string]interface{}, 2)
	map1["good_id"] = "hello111"
	map1["num"] = "world222"
	maps[0] = map1

	map2 := make(map[string]interface{}, 2)
	map2["good_id"] = "hello2222"
	map2["num"] = "world444"
	maps[1] = map2

	//return []byte
	str, err := json.Marshal(maps)

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("map to json", string(str))
}
