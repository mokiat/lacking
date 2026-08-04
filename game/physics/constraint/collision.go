package constraint

// func (s *PairCollision) ApplyImpulses(ctx solver.PairContext) {
// 	// Bounce solution
// 	pressureLambda := ctx.JacobianImpulseLambda(s.jacobian, 0.0, restitution)
// 	if pressureLambda > 0 {
// 		return // moving away
// 	}
// 	bounceSolution := ctx.JacobianImpulseSolution(s.jacobian, s.collisionDepth, restitution)

// 	// Friction solution
// 	primaryRadiusWS := dprec.QuatVec3Rotation(ctx.Target.Rotation(), s.primaryRadius)
// 	primaryPointVelocity := dprec.Vec3Sum(ctx.Target.LinearVelocity(), dprec.Vec3Cross(ctx.Target.AngularVelocity(), primaryRadiusWS))
// 	secondaryRadiusWS := dprec.QuatVec3Rotation(ctx.Source.Rotation(), s.secondaryRadius)
// 	secondaryPointVelocity := dprec.Vec3Sum(ctx.Source.LinearVelocity(), dprec.Vec3Cross(ctx.Source.AngularVelocity(), secondaryRadiusWS))
// 	deltaPointVelocity := dprec.Vec3Diff(primaryPointVelocity, secondaryPointVelocity)
// 	verticalVelocity := dprec.Vec3Prod(s.secondaryCollisionNormal, dprec.Vec3Dot(s.secondaryCollisionNormal, deltaPointVelocity))
// 	lateralVelocity := dprec.Vec3Diff(deltaPointVelocity, verticalVelocity)
// 	frictionSolution := solver.PairImpulse{}
// 	if lng := lateralVelocity.Length(); lng > solver.Epsilon {
// 		lateralDirection := dprec.UnitVec3(lateralVelocity)
// 		frictionJacobian := solver.PairJacobian{
// 			Target: solver.Jacobian{
// 				LinearSlope:  lateralDirection,
// 				AngularSlope: dprec.Vec3Cross(primaryRadiusWS, lateralDirection),
// 			},
// 			Source: solver.Jacobian{
// 				LinearSlope:  dprec.InverseVec3(lateralDirection),
// 				AngularSlope: dprec.Vec3Cross(lateralDirection, secondaryRadiusWS),
// 			},
// 		}
// 		frictionLambda := ctx.JacobianImpulseLambda(frictionJacobian, 0.0, 0.0)
// 		// TODO: Have friction coefficient configurable
// 		const frictionCoefficient = 0.9 // around 0.7 to 0.9 is realistic for dry asphalt
// 		maxFrictionLambda := pressureLambda * frictionCoefficient
// 		if -frictionLambda > -maxFrictionLambda {
// 			frictionLambda = maxFrictionLambda
// 		}
// 		frictionSolution = frictionJacobian.Impulse(frictionLambda)
// 	}

// 	// Note: Make sure to apply these as late as possible, otherwise you are
// 	// introducing noise that is picked up by subsequent calculations.
// 	ctx.Target.ApplyImpulse(bounceSolution.Target)
// 	ctx.Source.ApplyImpulse(bounceSolution.Source)
// 	ctx.Target.ApplyImpulse(frictionSolution.Target)
// 	ctx.Source.ApplyImpulse(frictionSolution.Source)
// }
