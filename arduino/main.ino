// Controller code for Arduino

#include <Wire.h>
#include <Servo.h>

#define SLAVE_ADDRESS       0x50

#define MAX_SERVO_ANGLE     150
#define MIN_SERVO_ANGLE     40
#define MAX_MOTOR_SPEED     150
#define MIN_MOTOR_SPEED     60
#define COLLISION_YES       8
#define COLLISION_NO        9
#define PROXIMITY_SENSOR    1 
#define SERVO_CONTROL_PIN   6
#define MOTOR_CONTROL_PIN   3
#define PROXIMITY_THRESHOLD 120
#define RESET_PIN           12
#define BREAK_TIME          750
#define DEFAULT_SERVO_ANGLE 90

char number =               0;
int control_word =          0;
int state =                 0;
int servo_angle =           90;
int motor_speed =           50;
int proximity =             0;
boolean set_servo_speed =   false;
boolean set_motor_speed =   false;
boolean i2c_active =        true;

Servo servo;

static const char SERVO_CONTROL = 'S';
static const char MOTOR_CONTROL = 'M';
static const char RESET = 'R';

void setup() {
    digitalWrite(RESET_PIN, HIGH);
    Wire.begin(SLAVE_ADDRESS);
    Wire.onReceive(receiveData);

    pinMode(RESET_PIN, OUTPUT);
    pinMode(SERVO_CONTROL_PIN, OUTPUT);
    pinMode(MOTOR_CONTROL_PIN, OUTPUT);
    pinMode(PROXIMITY_SENSOR, INPUT);
    pinMode(COLLISION_YES, OUTPUT);
    pinMode(COLLISION_NO,OUTPUT);

    digitalWrite(COLLISION_YES, HIGH);
    digitalWrite(COLLISION_NO, LOW);
    
    servo.attach(SERVO_CONTROL_PIN);
    servo.write(DEFAULT_SERVO_ANGLE);
}

void loop() {
    proximity = analogRead(PROXIMITY_SENSOR);
    if(proximity > PROXIMITY_THRESHOLD){
        digitalWrite(COLLISION_YES, HIGH);
        digitalWrite(COLLISION_NO, LOW);
        reset();
    }
    else{
        digitalWrite(COLLISION_YES, LOW);
        digitalWrite(COLLISION_NO, HIGH);
    }
    delay(20);
}

void receiveData(int byteCount) {
    while(Wire.available() && i2c_active) {
        control_word = Wire.read();

        if(set_servo_speed) {
            servo_angle = control_word;

            if(servo_angle > MAX_SERVO_ANGLE) {
              servo_angle = MAX_SERVO_ANGLE;
            } else if(servo_angle < MIN_SERVO_ANGLE) {
              servo_angle = MIN_SERVO_ANGLE;
            }

            set_servo_speed = false;

            setServoAngle(servo_angle);
        } else if(set_motor_speed) {
            motor_speed = control_word;

            if (motor_speed > MAX_MOTOR_SPEED) {
                motor_speed = MAX_MOTOR_SPEED;
            } else if (motor_speed < MIN_MOTOR_SPEED) {
                motor_speed = 0;
            }

            set_motor_speed = false;

            setMotorSpeed(motor_speed);
        } else if(char(control_word) == SERVO_CONTROL) {
            set_servo_speed = true;
        } else if(char(control_word) == MOTOR_CONTROL) {
            set_motor_speed = true;
        } else if(char(control_word) == RESET) {
            digitalWrite(RESET_PIN, LOW);
        }
    }
}

void setMotorSpeed(int motor_speed){
    analogWrite(MOTOR_CONTROL_PIN, motor_speed);
}

void setServoAngle(int angle){
    servo.write(angle);
}

void reset(){
    setMotorSpeed(20);
    setServoAngle(90);
    delay(BREAK_TIME);
    setMotorSpeed(0);
    i2c_active = false;
    while(1){}
}