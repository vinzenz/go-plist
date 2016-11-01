// Copyright 2016 Vinzenz Feenstra. All rights reserved.
// Use of this source code is governed by a BSD-2-clause
// license that can be found in the LICENSE file.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/vinzenz/go-plist"
)

const hardValid1Test = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Email</key>
	<string>user@email.com</string>
	<key>Name</key>
	<string>Üsér Diacriticà</string>
	<key>Signature</key>
	<data>
	RIhF/3CgyXzPg2wCQ5LShf6W9khtqPcqUDLAHcAZdOIcoeR7PoOHi15423kxq5jOh1lm
	cztBoUJFu8mB45MHE0jmmbRw3qK6FJz9Py2gi1XvGOgH3GW713OCvQBE7vfBj4ZriP0+
	FS18nLfrtM6Xp0mAd1la4DD4oh7d35dlYTY=
	</data>
	<key>lowercase key</key>
	<string>Keys should be sorted case-insensitive</string>
</dict>
</plist>`

func main() {
	x, e := plist.Read(bytes.NewReader([]byte(hardValid1Test)))
	if e != nil {
		fmt.Printf("ERROR: %s\n", e.Error())
	} else {
		json.NewEncoder(os.Stdout).Encode(x)
		json.NewEncoder(os.Stdout).Encode(x.Raw())
		x.Write(os.Stdout)
	}
}
