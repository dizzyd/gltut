package main

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
)

// CameraMovement type
type CameraMovement uint8

// CameraMovement consts
const (
	MoveForward CameraMovement = iota
	MoveBackward
	MoveLeft
	MoveRight
)

// Camera is a thing
type Camera struct {
	movementSpeed float64
	sensitivity   float64

	yaw   float64
	pitch float64

	position mgl32.Vec3
	front    mgl32.Vec3
	up       mgl32.Vec3
	worldUp  mgl32.Vec3
}

func newCamera() *Camera {
	return &Camera{
		movementSpeed: 3.0,
		sensitivity:   0.15,
		position:      mgl32.Vec3{0.0, 0.0, 3.0},
		up:            mgl32.Vec3{0.0, 1.0, 0.0},
		front:         mgl32.Vec3{0.0, 0.0, -1.0},
		yaw:           0.0,
		pitch:         0.0,
	}
}

func (c *Camera) viewMatrix() mgl32.Mat4 {
	return mgl32.LookAtV(c.position, c.position.Add(c.front), c.up)
}

func (c *Camera) processKeyboard(direction CameraMovement, deltaTime float32) {
	velocity := float32(c.movementSpeed) * deltaTime
	switch direction {
	case MoveForward:
		c.position = c.position.Add(c.front.Mul(velocity))
	case MoveBackward:
		c.position = c.position.Sub(c.front.Mul(velocity))
	case MoveLeft:
		c.position = c.position.Sub(c.front.Cross(c.up).Normalize().Mul(velocity))
	case MoveRight:
		c.position = c.position.Add(c.front.Cross(c.up).Normalize().Mul(velocity))
	}
}

func (c *Camera) processMousePos(xoffset, yoffset float64) {
	xoffset *= c.sensitivity
	yoffset *= c.sensitivity

	c.yaw += xoffset
	c.pitch += yoffset

	c.pitch = math.Max(-89.0, math.Min(89.0, c.pitch))

	c.updateVectors()
}

func (c *Camera) updateVectors() {
	yawR := mgl64.DegToRad(c.yaw)
	pitchR := mgl64.DegToRad(c.pitch)

	var front mgl32.Vec3
	front[0] = float32(math.Cos(pitchR) * math.Cos(yawR))
	front[1] = float32(math.Sin(pitchR))
	front[2] = float32(math.Cos(pitchR) * math.Sin(yawR))

	c.front = front.Normalize()
}
