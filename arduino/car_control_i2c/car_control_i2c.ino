#include <Wire.h>
#include <Servo.h>
#define SLAVE_ADDRESS 0x50
#define MAX_SERVO_ANGLE 150
#define MIN_SERVO_ANGLE 40
#define MAX_MOTOR_SPEED 170
#define MIN_MOTOR_SPEED 60
#define SERVO_CONTROL_PIN 6
#define MOTOR_CONTROL_PIN 3
static const char SERVO_CONTROL = 'S';
static const char MOTOR_CONTROL = 'M';

Servo servo;
char number = 0;
int control_word = 0;
int state = 0;
int servo_angle = 90;
int motor_speed = 50;
boolean set_servo_speed = false;
boolean set_motor_speed = false;

void setup() {
    Wire.begin(SLAVE_ADDRESS);
    Wire.onReceive(receiveData);

    pinMode(SERVO_CONTROL_PIN, OUTPUT);
    pinMode(MOTOR_CONTROL_PIN, OUTPUT);
    servo.attach(SERVO_CONTROL_PIN);
}

void loop() {

  delay(1000);
}

void receiveData(int byteCount){
    while(Wire.available()) {
        control_word = Wire.read();
        if(set_servo_speed){
            servo_angle = control_word;
            if(servo_angle > MAX_SERVO_ANGLE) {
              servo_angle = MAX_SERVO_ANGLE;
            } else if(servo_angle < MIN_SERVO_ANGLE) {
              servo_angle = MIN_SERVO_ANGLE;
            }
            set_servo_speed = false;
            setServoAngle(servo_angle);
        }
        else if(set_motor_speed){
            motor_speed = control_word;
            if (motor_speed > MAX_MOTOR_SPEED) {
                motor_speed = MAX_MOTOR_SPEED; 
            }
            else if (motor_speed < MIN_MOTOR_SPEED ) {
                motor_speed = MIN_MOTOR_SPEED; 
            }
            set_motor_speed = false;
            setMotorSpeed(motor_speed);
        }
        else if(char(control_word) == SERVO_CONTROL){
            set_servo_speed = true;
        }
        else if(char(control_word) == MOTOR_CONTROL){
            set_motor_speed = true;
        }
    }
}

void setMotorSpeed(int motor_speed){
    analogWrite(MOTOR_CONTROL_PIN, motor_speed);
}

void setServoAngle(int angle){
    servo.write(angle);
}


