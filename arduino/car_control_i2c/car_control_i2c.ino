#include <Wire.h>
#include <Servo.h>
#define SLAVE_ADDRESS 0x50

Servo servo;
char number = 0;
int control_word = 0;
int state = 0;
int servo_angle = 90;
int motor_speed = 50;
int motor_control_pin = 3;
int servo_control_pin = 6;
boolean set_servo_speed = false;
boolean set_motor_speed = false;

void setup() {
    Serial.begin(9600);      
     Wire.begin(SLAVE_ADDRESS);

     Wire.onReceive(receiveData);
    Wire.onRequest(sendData);

    Serial.println("Ready!");

    pinMode(servo_control_pin, OUTPUT);
    pinMode(motor_control_pin, OUTPUT);
    servo.attach(servo_control_pin);
}

void loop() {

  delay(1000);
}

void receiveData(int byteCount){
    while(Wire.available()) {
        control_word = Wire.read();
        if(set_servo_speed){
            servo_angle = control_word;
            if(servo_angle > 150) {
              servo_angle = 160;
            } else if(servo_angle < 30) {
              servo_angle = 30;
            }
            set_servo_speed = false;
            setServoAngle(servo_angle);
            Serial.print("S");
            Serial.println(servo_angle);
        }
        else if(set_motor_speed){
            motor_speed = control_word;
            if (motor_speed > 200) {
             motor_speed = 200; 
            }
//            else if (motor_speed < 50 ) {
//             motor_speed = 50; 
//            }
            set_motor_speed = false;
            setMotorSpeed(motor_speed);
            Serial.print('M');
            Serial.println(motor_speed);
        }
        else if(char(control_word) == 'S'){
            set_servo_speed = true;
        }
        else if(char(control_word) == 'M'){
            set_motor_speed = true;
        }
    }
}

void setMotorSpeed(int motor_speed){
    analogWrite(motor_control_pin, motor_speed);
}

void setServoAngle(int angle){
    servo.write(angle);
}

void sendData(){
    Wire.write(number);
}

