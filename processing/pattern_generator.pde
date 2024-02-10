// Test program for the PEmbroider library for Processing:
// Filling an image with the experimental SPINE hatch

import processing.embroider.*;
PEmbroiderGraphics E;
PImage myImage;

void setup() {
  noLoop(); 
  size (1500, 1500);

  E = new PEmbroiderGraphics(this, width, height);
  String outputFilePath = sketchPath("koala2.gcode");
  E.setPath(outputFilePath); 

  // The image should consist of white shapes on a black background. 
  // The ideal image is an exclusively black-and-white .PNG or .GIF.
  myImage = loadImage("koala2.png");
  
  int size_x = 360;
  int size_y = 360;
  float scalar = 2.25;
  
  myImage.resize((int)(size_x * scalar), (int)(size_y * scalar));
  
  E.beginDraw(); 
  E.clear();
  
  E.strokeMode(PEmbroiderGraphics.TANGENT);
  E.hatchMode(PEmbroiderGraphics.PARALLEL);
  //E.strokeWeight(1); 
  E.noStroke();

  PFont myFont = createFont("Helvetica-Bold", 200);
  

  E.strokeMode(PEmbroiderGraphics.TANGENT);
  E.noFill();
  E.hatchSpacing(3);
  E.strokeSpacing(10);
  E.stitchLength(40); 
  
  E.textFont(myFont);
  E.hatchMode(PEmbroiderGraphics.PARALLEL);
  E.fill(255); 
  E.noStroke();
  E.textSize(180);

  E.textAlign(LEFT, BOTTOM);
  //E.text("e's", 0,   200);
  
  //E.setStitch(5, 30, 0);
  //E.hatchMode(PEmbroiderGraphics.CROSS); 
  //E.HATCH_ANGLE = radians(30);
  //E.HATCH_ANGLE2 = radians(0);
  //E.hatchSpacing(4.0); 
  E.image(myImage, 250, 250);

  

  // Uncomment for the Alternative "Spine" rendering style:
  //PEmbroiderHatchSpine.setGraphics(EPASSTHROUGH_PREFIX, PASSTHROUGH_LENgra);
  //PEmbroiderHatchSpine.hatchSpineVF(myImage, 5);

  //-----------------------
   E.optimize();   // slow, but good and important
  E.visualize();  // 
  E.printStats(); //
   E.endDraw();    // write out the file
   save("PEmbroider_bitmap_image_2.png");
}


//--------------------------------------------
void draw() {
  ;
}
