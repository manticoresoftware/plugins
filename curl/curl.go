package main

import "C"
import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// Curl takes 1 string param with url, and returns str contents of resource by url.
// You can load it into the daemon with
//  CREATE FUNCTION curl RETURNS STRING SONAME 'curl.so';
// SELECT curl('https://yandex.ru/robots.txt');

// in init function we check only num of args (1) and the type (must be string)
//export curl_init
func curl_init(init *SPH_UDF_INIT, args *SPH_UDF_ARGS, errmsg *ERR_MSG) int32 {
	// check argument count
	if args.arg_count != 1 || args.arg_type(0) != SPH_UDF_TYPE_STRING {
		return errmsg.say("CURL() requires 1 string argument")
	}
	return 0
}

// here we execute provided action: load the resource and return it as allocated C string.
//export curl
func curl(init *SPH_UDF_INIT, args *SPH_UDF_ARGS, errf *ERR_FLAG) uintptr {

	url := args.stringval(0)

	// Get the data
	resp, _ := http.Get(url)
	defer func() { _ = resp.Body.Close() }()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return args.return_string(fmt.Sprintf("Bad status: %s", resp.Status))
	}

	// let's check content-type and avoid anything, but text
	contentType := resp.Header.Get("Content-Type")
	parts := strings.Split(contentType, "/")
	if len(parts) >= 1 && parts[0] == "text" {
		// retrieve whole body
		text, _ := ioutil.ReadAll(resp.Body)
		return args.return_string(string(text))
	}

	return args.return_string("Content type: " + contentType + ", will NOT download")
}
