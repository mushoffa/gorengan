package client

import "testing"

func TestNewFiber_ReturnsTheRequestedVersion(t *testing.T) {
	// Given
	fiberFactory := NewFiber()

	// When
	v2 := fiberFactory.V2()
	v3 := fiberFactory.V3()

	// Then
	if _, ok := v2.(*fiberV2); !ok {
		t.Errorf("V2() = %T, want *fiberV2", v2)
	}
	if _, ok := v3.(*fiberV3); !ok {
		t.Errorf("V3() = %T, want *fiberV3", v3)
	}
}

func TestFactory_Fiber_ReturnsTheRequestedVersion(t *testing.T) {
	// Given
	fiberFactory := New().Fiber()

	// When
	v2 := fiberFactory.V2()
	v3 := fiberFactory.V3()

	// Then
	if _, ok := v2.(*fiberV2); !ok {
		t.Errorf("V2() = %T, want *fiberV2", v2)
	}
	if _, ok := v3.(*fiberV3); !ok {
		t.Errorf("V3() = %T, want *fiberV3", v3)
	}
}

func TestFiberFactory_EachCallBuildsANewClient(t *testing.T) {
	// Given
	fiberFactory := NewFiber()

	// When
	firstV2, secondV2 := fiberFactory.V2(), fiberFactory.V2()
	firstV3, secondV3 := fiberFactory.V3(), fiberFactory.V3()

	// Then
	if firstV2 == secondV2 {
		t.Error("V2() returned the same client twice, want a new one per call")
	}
	if firstV3 == secondV3 {
		t.Error("V3() returned the same client twice, want a new one per call")
	}
}
