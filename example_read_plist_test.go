// Copyright 2016 Vinzenz Feenstra. All rights reserved.
// Use of this source code is governed by a BSD-2-clause
// license that can be found in the LICENSE file.
package plist_test

import (
	"bytes"
	"fmt"
	"time"

	"github.com/vinzenz/go-plist"
)

const exampleReadPlistData = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>Email</key>
		<string>user@example.com</string>
		<key>Name</key>
		<string>Üsér Diacriticà</string>
		<key>Signature</key>
		<data>
		RIhF/3CgyXzPg2wCQ5LShf6W9khtqPcqUDLAHcAZdOIcoeR7PoOHi15423kxq5jOh1lm
		cztBoUJFu8mB45MHE0jmmbRw3qK6FJz9Py2gi1XvGOgH3GW713OCvQBE7vfBj4ZriP0+
		FS18nLfrtM6Xp0mAd1la4DD4oh7d35dlYTY=
		</data>
		<key>Some integer</key>
		<integer>-131383</integer>
		<key>Some floating point number</key>
		<real>-14242424.342</real>
		<key>Another floating point number</key>
		<real>-2.0e+04</real>
		<key>Generated</key>
	    <date>2016-11-01T08:46:41Z</date>
	</dict>
</plist>`

func ExampleRead() {
	if parsed, err := plist.Read(bytes.NewReader([]byte(exampleReadPlistData))); err == nil {
		rawData := parsed.Raw().(map[string]interface{})
		fmt.Printf("Name:      %s\n", rawData["Name"].(string))
		fmt.Printf("Email: 	   %s\n", rawData["Email"].(string))
		fmt.Printf("A number:  %d\n", rawData["Some integer"].(int64))
		fmt.Printf("Generated: %s\n", rawData["Email"].(time.Time).String())
	} else {
		fmt.Printf("Failed to parse example data: %s\n", err.Error())
	}
}
