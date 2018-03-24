%module rf24
%{
#include "RF24.h"
#include "RF24_config.h"
#include "nRF24L01.h"
#include "utility/includes.h"
%}

%include <typemaps.i>
%include "std_string.i"
%include "std_vector.i"

// This will create 2 wrapped types in Go called
// "StringVector" and "ByteVector" for their respective
// types.
namespace std {
   %template(StringVector) vector<string>;
   %template(ByteVector) vector<char>;
}

%include "RF24.h"
%include "RF24_config.h"
%include "nRF24L01.h"
%include "utility/includes.h"
