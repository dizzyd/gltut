#version 410 core

layout (location = 0) in vec3 position;
layout (location = 1) in vec3 color;
layout (location = 2) in vec2 texture;

uniform float horizOffset; 

out vec3 ourColor;
out vec2 texCoord;

void main() {
    gl_Position = vec4(position.x + horizOffset, position.y, position.z, 1.0);
    ourColor = color;
    texCoord = texture;
}