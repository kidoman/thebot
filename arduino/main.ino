// Controller code for Arduino

#include <Wire.h>
#include <Servo.h>

//Address at which arduino is connected to I2C bus as a slave
#define SLAVE_ADDRESS               0x50

//Servo constants
#define MAX_SERVO_ANGLE             150
#define MIN_SERVO_ANGLE             40
#define DEFAULT_SERVO_ANGLE         90

//Motor Parameters
#define MAX_MOTOR_SPEED             150
#define MIN_MOTOR_SPEED             60

//For Proximity Sensing
#define COLLISION_YES               8
#define COLLISION_NO                9
#define MIN_COLLISION_THRESHOLD     120

//IO Pins
#define PROXIMITY_SENSOR            1
#define SERVO_CONTROL_PIN           6
#define MOTOR_CONTROL_PIN           3
#define RESET_PIN                   12
#define PROXIMITY_THRESHOLD_CONTROL 2

//Misc
#define BREAK_TIME                  750

char number =                       0;
int control_word =                  0;
int state =                         0;
int servo_angle =                   90;
int motor_speed =                   50;
int proximity =                     0;
int proximity_control =             0;
int manual_proximity_threshold =    0;
boolean set_servo_speed =           false;
boolean set_motor_speed =           false;
boolean on_reset =                  true;

Servo servo;

static const char SERVO_CONTROL =   'S';
static const char MOTOR_CONTROL =   'M';
static const char RESET =           'R';

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
    pinMode(PROXIMITY_THRESHOLD_CONTROL, INPUT);

    digitalWrite(COLLISION_YES, HIGH);
    digitalWrite(COLLISION_NO, LOW);

    servo.attach(SERVO_CONTROL_PIN);
    servo.write(DEFAULT_SERVO_ANGLE);
}

void loop() {
    manual_proximity_threshold = analogRead(PROXIMITY_THRESHOLD_CONTROL);
    proximity = analogRead(PROXIMITY_SENSOR);
    if (proximity > MIN_COLLISION_THRESHOLD || proximity > manual_proximity_threshold) {
        turn_on_proximity_warning();
        reset();
    } else {
        turn_off_proximity_warning();
    }
    delay(20);
}

void receiveData(int byteCount) {
    while(Wire.available()) {
        control_word = Wire.read();

        if (set_servo_speed) {
            servo_angle = control_word;

            if (servo_angle > MAX_SERVO_ANGLE) {
              servo_angle = MAX_SERVO_ANGLE;
            } else if (servo_angle < MIN_SERVO_ANGLE) {
              servo_angle = MIN_SERVO_ANGLE;
            }

            set_servo_speed = false;

            setServoAngle(servo_angle);
        } else if (set_motor_speed) {
            motor_speed = control_word;

            if (motor_speed > MAX_MOTOR_SPEED) {
                motor_speed = MAX_MOTOR_SPEED;
            } else if (motor_speed < MIN_MOTOR_SPEED) {
                motor_speed = 0;
            }

            set_motor_speed = false;

            setMotorSpeed(motor_speed);
        } else if (char(control_word) == SERVO_CONTROL && !(on_reset)) {
            set_servo_speed = true;
        } else if (char(control_word) == MOTOR_CONTROL && !(on_reset)) {
            set_motor_speed = true;
        } else if (char(control_word) == RESET) {
            digitalWrite(RESET_PIN, LOW);
            set_servo_speed = false;
            set_motor_speed = false;
        }
    }
}

void setMotorSpeed(int motor_speed) {
    analogWrite(MOTOR_CONTROL_PIN, motor_speed);
}

void setServoAngle(int angle) {
    servo.write(angle);
}

void turn_on_proximity_warning() {
    digitalWrite(COLLISION_YES, HIGH);
    digitalWrite(COLLISION_NO, LOW);
}

void turn_off_proximity_warning() {
    digitalWrite(COLLISION_YES, LOW);
    digitalWrite(COLLISION_NO, HIGH);
}

void reset() {
    setMotorSpeed(20);
    setServoAngle(90);
    delay(BREAK_TIME);
    setMotorSpeed(0);
    on_reset = false;
    while(true) {}
}
