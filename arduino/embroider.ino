/*********
  Rui Santos
  Complete project details at http://randomnerdtutorials.com  
*********/
#include <SoftwareSerial.h>

// TCS230 or TCS3200 pins wiring to Arduino
#define S0 4
#define S1 5
#define S2 6
#define S3 7
#define sensorOut 8

// Stores frequency read by the photodiodes
int redFrequency = 0;
int greenFrequency = 0;
int blueFrequency = 0;

// Stores the red. green and blue colors
int redColor = 0;
int greenColor = 0;
int blueColor = 0;

bool running = false;
unsigned long timeLastPulse = 0;

unsigned long min_pulses = 20;
unsigned long pulse_count = 0;

String inputString = "";      // a String to hold incoming data
bool stringComplete = false;  // whether the string is complete
SoftwareSerial  gSerialPassthrough(9, 10);
const char* PASSTHROUGH_PREFIX = "g:";
size_t PASSTHROUGH_LEN = 2;

void setup() {
  // Setting the outputs
  pinMode(S0, OUTPUT);
  pinMode(S1, OUTPUT);
  pinMode(S2, OUTPUT);
  pinMode(S3, OUTPUT);

  pinMode(3, OUTPUT);          // sets the digital pin 13 as output
  
  // Setting the sensorOut as an input
  pinMode(sensorOut, INPUT);
  
  // Setting frequency scaling to 20%
  digitalWrite(S0,HIGH);
  digitalWrite(S1,LOW);

  digitalWrite(S2,LOW);
  digitalWrite(S3,LOW);
  
  // Begins serial communication
  Serial.begin(115200);
  gSerialPassthrough.begin(57600);

  inputString.reserve(200);
}

void serialEvent() {
  while (Serial.available()) {
    // get the new byte:
    char inChar = (char)Serial.read();
    // add it to the inputString:
    inputString += inChar;
    // if the incoming character is a newline, set a flag so the main loop can
    // do something about it:
    if (inChar == '\n') {
      stringComplete = true;
    }
  }
}

bool safe() {
  // // Reading the output frequency
  redFrequency = pulseIn(sensorOut, LOW);
  // Remaping the value of the RED (R) frequency from 0 to 255
  // You must replace with your own values. Here's an example: 
  redColor = map(redFrequency, 50, 100, 255,0);
  return redColor < 50;
}

bool canPulse() {
  return 100 < millis() - timeLastPulse;
}

void pulse(int length) {
  digitalWrite(3, HIGH);
  delay(length);
  digitalWrite(3, LOW);
  timeLastPulse = millis();
}

void loop() {
  gSerialPassthrough.listen();
  if (gSerialPassthrough.available())
  {
    char byte = gSerialPassthrough.read();
    // Send header with every byte
    // Serial.write(PASSTHROUGH_PREFIX, PASSTHROUGH_LEN);
    Serial.write(byte);
  }
  if (stringComplete)
  {
    inputString.trim();
    const char* str = inputString.c_str();
    if (memcmp(str, PASSTHROUGH_PREFIX, PASSTHROUGH_LEN) == 0)
    {
      // Passthrough to other board
      gSerialPassthrough.write(str + PASSTHROUGH_LEN);
      gSerialPassthrough.write("\r\n");
    }
    else if (inputString == "PULSE")
    {
      pulse_count = 0;
      Serial.println("ACK");
      pulse(120);
      running = true;
    }
    else if (inputString == "?")
    {
      Serial.println(running ? "RUNNING" : "IDLE");
    }
    else
    {
      Serial.println("NAK");
    }

    inputString = "";
    stringComplete = false;
  }

  if (running && canPulse())
  {
    if (!safe())
    {
      pulse_count += 1;
      pulse(18);
    } else
    {
      running = false;
    }
  }
}