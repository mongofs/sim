/*
 * Copyright 2022 steven
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *    http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"
	"strconv"
)

type Color int

// Foreground text colors
const (
	FgBlack Color = iota + 30
	FgRed
	FgGreen
	FgYellow
	FgBlue
	FgMagenta
	FgCyan
	FgWhite
)

// Foreground Hi-Intensity text colors
const (
	FgHiBlack Color = iota + 90
	FgHiRed
	FgHiGreen
	FgHiYellow
	FgHiBlue
	FgHiMagenta
	FgHiCyan
	FgHiWhite
)

// Colorize a string based on given color.
func PrintWithColor(s string, c Color)  {
	 fmt.Printf("\033[1;%s;40m%s\033[0m\n", strconv.Itoa(int(c)), s)
}

// Colorize a string based on given color.
func PrintWithColorAndSpace(s string, c Color,front,back int)  {
	base := fmt.Sprintf("\033[1;%s;40m%s\033[0m\n", strconv.Itoa(int(c)), s)
	for front > 0 {
		base = fmt.Sprintf("\n%s",base)
		front --
	}
	for back > 0 {
		base = fmt.Sprintf("%s\n",base)
		back --
	}
	fmt.Printf(base)
}