package html

/*
#cgo CFLAGS: -I../../../clibs/include/libxml2
#cgo LDFLAGS: -lxml2 -L../../../clibs/lib

#include <libxml/HTMLtree.h>
#include <libxml/HTMLparser.h>
#include "helper.h"
*/
import "C"

import (
	"errors"
	. "gokogiri/util"
	"gokogiri/xml"
	"unsafe"

	"time"
)

//xml parse option
const (
	HTML_PARSE_RECOVER   = 1 << 0  /* Relaxed parsing */
	HTML_PARSE_NODEFDTD  = 1 << 2  /* do not default a doctype if not found */
	HTML_PARSE_NOERROR   = 1 << 5  /* suppress error reports */
	HTML_PARSE_NOWARNING = 1 << 6  /* suppress warning reports */
	HTML_PARSE_PEDANTIC  = 1 << 7  /* pedantic error reporting */
	HTML_PARSE_NOBLANKS  = 1 << 8  /* remove blank nodes */
	HTML_PARSE_NONET     = 1 << 11 /* Forbid network access */
	HTML_PARSE_NOIMPLIED = 1 << 13 /* Do not add implied html/body... elements */
	HTML_PARSE_COMPACT   = 1 << 16 /* compact small text nodes */
)

const EmptyHtmlDoc = ""

//default parsing option: relax parsing
var DefaultParseOption = HTML_PARSE_RECOVER |
	HTML_PARSE_NONET |
	HTML_PARSE_NOERROR |
	HTML_PARSE_NOWARNING

type HtmlDocument struct {
	*xml.XmlDocument

	StartTime int64
}

//default encoding in byte slice
var DefaultEncodingBytes = []byte(xml.DefaultEncoding)
var emptyHtmlDocBytes = []byte(EmptyHtmlDoc)

var ErrSetMetaEncoding = errors.New("Set Meta Encoding failed")
var ERR_FAILED_TO_PARSE_HTML = errors.New("failed to parse html input")
var emptyStringBytes = []byte{0}

//create a document
func NewDocument(p unsafe.Pointer, contentLen int, inEncoding, outEncoding []byte) (doc *HtmlDocument) {
	doc = &HtmlDocument{}
	doc.XmlDocument = xml.NewDocument(p, contentLen, inEncoding, outEncoding)
	doc.Me = doc
	node := doc.Node.(*xml.XmlNode)
	node.Document = doc
	return
}

//parse a string to document
func Parse(content, inEncoding, url []byte, options int, outEncoding []byte) (doc *HtmlDocument, err error) {
	startTime := time.Now()

	inEncoding = AppendCStringTerminator(inEncoding)
	outEncoding = AppendCStringTerminator(outEncoding)

	var docPtr *C.xmlDoc
	contentLen := len(content)

	if contentLen > 0 {
		var contentPtr, urlPtr, encodingPtr unsafe.Pointer

		contentPtr = unsafe.Pointer(&content[0])
		if len(url) > 0 {
			url = AppendCStringTerminator(url)
			urlPtr = unsafe.Pointer(&url[0])
		}
		if len(inEncoding) > 0 {
			encodingPtr = unsafe.Pointer(&inEncoding[0])
		}

		docPtr = C.htmlParse(contentPtr, C.int(contentLen), urlPtr, encodingPtr, C.int(options), nil, 0)

		if docPtr == nil {
			err = ERR_FAILED_TO_PARSE_HTML
		} else {
			doc = NewDocument(unsafe.Pointer(docPtr), contentLen, inEncoding, outEncoding)
		}
	}
	if docPtr == nil {
		doc = CreateEmptyDocument(inEncoding, outEncoding)
	}

	if xml.ProfilingData["HtmlDocument.Parse"] == nil {
		xml.ProfilingData["HtmlDocument.Parse"] = &xml.CountAndSum{ 1, time.Since(startTime) }
	} else {
		xml.ProfilingData["HtmlDocument.Parse"].Count++
		xml.ProfilingData["HtmlDocument.Parse"].Sum += time.Since(startTime)
	}
	return
}

func CreateEmptyDocument(inEncoding, outEncoding []byte) (doc *HtmlDocument) {
	C.xmlInitParser()
	docPtr := C.htmlNewDoc(nil, nil)
	doc = NewDocument(unsafe.Pointer(docPtr), 0, inEncoding, outEncoding)
	return
}

func (document *HtmlDocument) ParseFragment(input, url []byte, options int) (fragment *xml.DocumentFragment, err error) {
	root := document.Root()
	if root == nil {
		fragment, err = parsefragment(document, nil, input, url, options)
	} else {
		fragment, err = parsefragment(document, root.XmlNode, input, url, options)
	}
	return
}

func (doc *HtmlDocument) MetaEncoding() string {
	metaEncodingXmlCharPtr := C.htmlGetMetaEncoding((*C.xmlDoc)(doc.DocPtr()))
	return C.GoString((*C.char)(unsafe.Pointer(metaEncodingXmlCharPtr)))
}

func (doc *HtmlDocument) SetMetaEncoding(encoding string) (err error) {
	var encodingPtr unsafe.Pointer = nil
	if len(encoding) > 0 {
		encodingBytes := AppendCStringTerminator([]byte(encoding))
		encodingPtr = unsafe.Pointer(&encodingBytes[0])
	}
	ret := int(C.htmlSetMetaEncoding((*C.xmlDoc)(doc.DocPtr()), (*C.xmlChar)(encodingPtr)))
	if ret == -1 {
		err = ErrSetMetaEncoding
	}
	return
}
