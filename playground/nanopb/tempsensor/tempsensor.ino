#include <pb_decode.h>
#include <pb_common.h>
#include <pb.h>
#include <pb_encode.h>

#include "temp.pb.h"

void PrintHex8(uint8_t *data, uint8_t length) // prints 8-bit data in hex with leading zeroes
{
     char tmp[16];
       for (int i=0; i<length; i++) { 
         sprintf(tmp, "0x%.2X",data[i]); 
         Serial.print(tmp); Serial.print(" ");
       }
}


void setup(){
  Serial.begin(115200);
  Serial.println(F("RF24/examples/GettingStarted"));
}

void loop() {

  Serial.println("creating  TempEvent...");
//  float hum = dht.readHumidity();
//  float tmp = dht.readTemperature();
//  float hiCel = dht.computeHeatIndex(tmp, hum, false);
    
  pb_TempEvent temp = pb_TempEvent_init_zero;
  temp.deviceId = 12;
  temp.eventId = 100;
  temp.humidity = 43.43;
  temp.tempCel = 12.3;
  temp.heatIdxCel = 33.33;

//  printjunk();
  sendTemp(temp);
     delay(1000);
}
//
//void printjunk(){
//  
//  uint8_t ByteData[5]={0x01, 0x0F, 0x10, 0x11, 0xFF};
//  Serial.print("With uint8_t array:  "); PrintHex8(ByteData,5); Serial.print("\n"); 
//  Serial.println("");
//}

void sendTemp(pb_TempEvent e) {
  uint8_t buffer[128];
  memset(buffer, 0, sizeof(buffer));
  pb_ostream_t stream = pb_ostream_from_buffer(buffer, sizeof(buffer));
  
  if (!pb_encode(&stream, pb_TempEvent_fields, &e)){
    Serial.println("failed to encode temp proto");
    return;
  }
  //client.write(buffer, stream.bytes_written);
  Serial.print(F("Sent "));
//  Serial.print(buffer);
  Serial.print("With uint8_t array:  "); PrintHex8(buffer,128); Serial.print("\n"); 
  Serial.println("");
}

