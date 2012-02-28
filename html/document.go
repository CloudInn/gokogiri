package html

/*
#cgo pkg-config: libxml-2.0

#include <libxml/HTMLtree.h>
#include <libxml/HTMLparser.h>
#include "chelper.h"
*/
import "C"

import (
	"unsafe"
	//"errors"
	"gokogiri/xml"
)

//xml parse option
const (
    HTML_PARSE_RECOVER  = 1<<0 	/* Relaxed parsing */
    HTML_PARSE_NODEFDTD = 1<<2 	/* do not default a doctype if not found */
    HTML_PARSE_NOERROR	= 1<<5	/* suppress error reports */
    HTML_PARSE_NOWARNING= 1<<6	/* suppress warning reports */
    HTML_PARSE_PEDANTIC	= 1<<7	/* pedantic error reporting */
    HTML_PARSE_NOBLANKS	= 1<<8	/* remove blank nodes */
    HTML_PARSE_NONET	= 1<<11 /* Forbid network access */
    HTML_PARSE_NOIMPLIED= 1<<13 /* Do not add implied html/body... elements */
    HTML_PARSE_COMPACT  = 1<<16 /* compact small text nodes */
)

//default parsing option: relax parsing
var DefaultParseOption = 	HTML_PARSE_RECOVER|
							HTML_PARSE_NONET|
							HTML_PARSE_NOERROR|
							HTML_PARSE_NOWARNING

type Document struct {
	*xml.Document
}

//default encoding in byte slice
var DefaultEncodingBytes = []byte(xml.DefaultEncoding)

//create a document
func NewDocument(p unsafe.Pointer, encoding []byte, buffer []byte) (doc *Document) {
	doc = &Document{}
	doc.Document = xml.NewDocument(p, encoding, buffer)
	return
}

//parse a string to document
func Parse(content, url, encoding []byte, options int) (doc *Document, err error) {
	var docPtr *C.xmlDoc
	contentLen := len(content)
	
	if contentLen > 0 {
		var contentPtr, urlPtr, encodingPtr unsafe.Pointer
		
		contentPtr   = unsafe.Pointer(&content[0])
		if len(url) > 0      { urlPtr       = unsafe.Pointer(&url[0]) }
		if len(encoding) > 0 { encodingPtr  = unsafe.Pointer(&encoding[0]) }
		
		docPtr = C.html_native_parse(contentPtr, C.int(contentLen), urlPtr, encodingPtr, C.int(options), nil, 0)
	}
	if docPtr == nil {
		//why does newEmptyXmlDoc NOT call xmlInitParser like other parse functions?
		C.xmlInitParser();
		docPtr = C.htmlNewDoc(nil, nil)
	}
	
	doc = NewDocument(unsafe.Pointer(docPtr), encoding, nil)
	return
}

func (document *Document) String() string {
	return document.ToHtml()
}