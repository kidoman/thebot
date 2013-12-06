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

#define SERVO_CONTROL               'S'
#define MOTOR_CONTROL               'M'
#define RESET                       'R'

#define SERIAN_ON                   true

boolean set_servo_angle =           false;
boolean set_motor_speed =           false;
boolean on_reset =                  false;

Servo servo;

void setup() {
    if (SERIAN_ON) {
        Serial.begin(9600);
    }
    
    Serial.println("Startup...");

    on_reset = false;
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

    turn_off_proximity_warning();

    servo.attach(SERVO_CONTROL_PIN);
    servo.write(DEFAULT_SERVO_ANGLE);
}

void loop() {
    int proximity = analogRead(PROXIMITY_SENSOR);
    if (proximity > MIN_COLLISION_THRESHOLD) {
        turn_on_proximity_warning();
        halt();
    }
    delay(20);
}

void receiveData(int byteCount) {
    while(Wire.available()) {
        int control_word = Wire.read();

        if (set_servo_angle) {
            int servo_angle = control_word;

            if (servo_angle > MAX_SERVO_ANGLE) {
              servo_angle = MAX_SERVO_ANGLE;
            } else if (servo_angle < MIN_SERVO_ANGLE) {
              servo_angle = MIN_SERVO_ANGLE;
            }

            set_servo_angle = false;

            if (!on_reset) {
                setServoAngle(servo_angle);    
            }
            
        } 
        else if (set_motor_speed) {
            int motor_speed = control_word;

            if (motor_speed > MAX_MOTOR_SPEED) {
                motor_speed = MAX_MOTOR_SPEED;
            } else if (motor_speed < MIN_MOTOR_SPEED) {
                motor_speed = 0;
            }

            set_motor_speed = false;

            if (!on_reset) {
                setMotorSpeed(motor_speed);    
            }
            
        } 
        else if (char(control_word) == SERVO_CONTROL) {
            set_servo_angle = true;
            set_motor_speed = false;
        } 
        else if (char(control_word) == MOTOR_CONTROL) {
            set_motor_speed = true;
            set_servo_angle = false;
        } 
        else if (char(control_word) == RESET) {
            digitalWrite(RESET_PIN, LOW);
            set_servo_angle = false;
            set_motor_speed = false;
        }
    }
}

void setMotorSpeed(int motor_speed) {
    Serial.print("Setting motor speed to");
    Serial.println(motor_speed);
    analogWrite(MOTOR_CONTROL_PIN, motor_speed);
}

void setServoAngle(int angle) {
    Serial.print("Setting angle to");
    Serial.println(angle);
    servo.write(angle);
}

void turn_on_proximity_warning() {
    Serial.println("Turning on proximity warning...");
    digitalWrite(COLLISION_YES, HIGH);
    digitalWrite(COLLISION_NO, LOW);
}

void turn_off_proximity_warning() {
    Serial.println("Turning off proximity warning...");
    digitalWrite(COLLISION_YES, LOW);
    digitalWrite(COLLISION_NO, HIGH);
}

void halt() {
    Serial.println("Halting...");
    setMotorSpeed(20);
    setServoAngle(90);
    delay(BREAK_TIME);
    setMotorSpeed(0);
    on_reset = true;
    Serial.println("Halted...");
    while(true) {}
}
