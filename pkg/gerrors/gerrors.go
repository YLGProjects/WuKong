/**
 *MIT License
 *
 *Copyright (c) 2025 ylgeeker
 *
 *Permission is hereby granted, free of charge, to any person obtaining a copy
 *of this software and associated documentation files (the "Software"), to deal
 *in the Software without restriction, including without limitation the rights
 *to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *copies of the Software, and to permit persons to whom the Software is
 *furnished to do so, subject to the following conditions:
 *
 *copies or substantial portions of the Software.
 *
 *THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *SOFTWARE.
**/

package gerrors

import "fmt"

// Code global error code
type Code int

const (
	// System Errors (1000-1999)
	Success Code = iota + 1000
	InternalServer
	Timeout
	InvalidNetAddress

	// User Errors (2000-2999)
	UserNotFound = iota + 2000
	UserDisabled
	UserAlreadyExists
	InvalidConfig
)

// GError global error type
type GError struct {
	code    Code
	message string
}

// New create a new GError object
func New(c Code, msg string) *GError {
	return &GError{code: c, message: msg}
}

// Code only return error code
func (g *GError) Code() Code {
	return g.code
}

// Message only return message
func (g *GError) Message() string {
	return g.message
}

// Error error interface method
func (g *GError) Error() string {
	return fmt.Sprintf("code:%d, errmsg:%s", g.code, g.message)
}
