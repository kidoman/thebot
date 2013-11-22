#include <Servo.h>

Servo servo;
int bin_0 = 4;
int bin_1 = 5;
int bin_2 = 6;
int bin_3 = 7;
int bin_4 = 8;
int servo_angle = 90;
int control_val = 0;
int motor_pin = 11;
int motor_speed = 0;
int motor_control_pin = 0;
int motor_control_val = 0;

void setup(){
  Serial.begin(9600);
  pinMode(bin_0, INPUT);
  pinMode(bin_1, INPUT);
  pinMode(bin_2, INPUT);
  pinMode(bin_3, INPUT);
  pinMode(bin_4, INPUT);
  pinMode(motor_control_pin, INPUT);
  pinMode(motor_pin, OUTPUT);
  servo.attach(9);
}

void loop(){
  motor_control_val = analogRead(motor_control_pin);
  motor_speed = map(motor_control_val, 0, 1023, 0, 254);
  analogWrite(motor_pin, motor_speed);
  
  control_val = 0;
  digitalRead(bin_0) == HIGH ? control_val=1 : control_val=0;
  digitalRead(bin_1) == HIGH ? control_val+=2 : control_val+=0;
  digitalRead(bin_2) == HIGH ? control_val+=4 : control_val+=0;
  digitalRead(bin_3) == HIGH ? control_val+=8 : control_val+=0;
  digitalRead(bin_4) == HIGH ? control_val+=16 : control_val+=0;
  
  if (control_val < 8) {
    control_val = 8;
  }
  else if (control_val > 20) {
    control_val = 20;
  }
  
  servo_angle = map(control_val, 0, 31, 0, 179);
  servo.write(servo_angle);
  Serial.println("control_val");
  Serial.println(control_val, DEC);
  Serial.println("servo_angle");
  Serial.println(servo_angle, DEC);
}
