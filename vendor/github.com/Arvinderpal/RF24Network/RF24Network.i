%module RF24Network
%{
#include "RF24Network.h"
#include "RF24Network_config.h"
#include "Sync.h"
%}

#define RF24_LINUX
%include <typemaps.i>
%include "stdint.i"
%include "std_string.i"
%include "std_vector.i"

// This will create 2 wrapped types in Go called
// "StringVector" and "ByteVector" for their respective
// types.
namespace std {
   %template(StringVector) vector<string>;
   %template(ByteVector) vector<char>;
}

%include "RF24Network.h"
%include "RF24Network_config.h"
%include "Sync.h"
