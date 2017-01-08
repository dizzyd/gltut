package main

import "fmt"
import "github.com/go-gl/mathgl/mgl32"

func nomain() {
	vec := mgl32.Vec4{1.0, 0.0, 0.0, 1.0}
	trans := mgl32.Translate3D(1.0, 1.0, 0.0)
	vec = trans.Mul4x1(vec)
	fmt.Printf("x:%4.2f y:%4.2f z:%4.2f\n", vec.X(), vec.Y(), vec.Z())

	t := mgl32.HomogRotate3D(mgl32.DegToRad(90.0), mgl32.Vec3{0.0, 0.0, 1.0})
	t = t.Mul4(mgl32.Scale3D(0.5, 0.5, 0.5))
	fmt.Printf("%+v\n", t)

	// trans = glm::rotate(trans, 90.0f, glm::vec3(0.0, 0.0, 1.0));
	// trans = glm::scale(trans, glm::vec3(0.5, 0.5, 0.5));
}
