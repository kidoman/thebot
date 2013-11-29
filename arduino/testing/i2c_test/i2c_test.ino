#include <Wire.h>

#define SLAVE_ADDRESS 0x50
char number = 0;
int state = 0;
int motor_control_pin = 3;

void setup() {
    Serial.begin(9600);      
    // initialize i2c as slave
    Wire.begin(SLAVE_ADDRESS);

    // define callbacks for i2c communication
    Wire.onReceive(receiveData);
    Wire.onRequest(sendData);

    Serial.println("Ready!");
}

void loop() {
    delay(500);
}

void receiveData(int byteCount){

    while(Wire.available()) {
        number = Wire.read();
        Serial.print("data received: ");
        Serial.println(number);
     }
}

void sendData(){
    Wire.write(number);
}

