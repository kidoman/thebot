#include <Wire.h>
#include <math.h>

#define MAG_REG 0x1E

void void setup()
{
	Serial.begin(9600);
	Wire.begin();

	Wire.beginTransmission(MAG_REG);
	Wire.write(0x00); //address
	Wire.write(0x14); //data
	Wire.endTransmission();

	Wire.beginTransmission(MAG_REG);
	Wire.write(0x02); //address
	Wire.write(0x00); //data
	Wire.endTransmission();
}


void loop()
{
	int mag_values[3];
	Wire.beginTransmission(MAG_REG);
  	Wire.write(0X03);
  	Wire.endTransmission();
  	Wire.requestFrom(MAG_REG, 6);
	for (int i=0; i<3; i++)
    	mag_values[i] = (Wire.read() << 8) | Wire.read();	

    float heading = 180*((atan2(mag_values[1], mag_values[0]))/3.14)

    if (heading <0)
    	heading += 360;

    Serial.print("X: ");
    Serial.print(mag_values[0]);
    Serial.print("  Y: ");
    Serial.print(mag_values[1]);
    Serial.print("  Z: ");
    Serial.print(mag_values[2]);    
    Serial.print("  Heading: ");
    Serial.println(heading);
    delay(200);
}