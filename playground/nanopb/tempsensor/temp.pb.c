/* Automatically generated nanopb constant definitions */
/* Generated by nanopb-0.4.0-dev at Sun Oct  7 11:41:22 2018. */

#include "temp.pb.h"

/* @@protoc_insertion_point(includes) */
#if PB_PROTO_HEADER_VERSION != 30
#error Regenerate this file with the current version of nanopb generator.
#endif



const pb_field_t pb_TempEvent_fields[6] = {
    PB_FIELD(  1, INT32   , REQUIRED, STATIC  , FIRST, pb_TempEvent, deviceId, deviceId, 0),
    PB_FIELD(  2, INT32   , REQUIRED, STATIC  , OTHER, pb_TempEvent, eventId, deviceId, 0),
    PB_FIELD(  3, FLOAT   , REQUIRED, STATIC  , OTHER, pb_TempEvent, humidity, eventId, 0),
    PB_FIELD(  4, FLOAT   , REQUIRED, STATIC  , OTHER, pb_TempEvent, tempCel, humidity, 0),
    PB_FIELD(  5, FLOAT   , REQUIRED, STATIC  , OTHER, pb_TempEvent, heatIdxCel, tempCel, 0),
    PB_LAST_FIELD
};


/* @@protoc_insertion_point(eof) */
