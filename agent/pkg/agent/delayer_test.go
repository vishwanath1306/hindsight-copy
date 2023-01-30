package agent

import (
	"testing"
	"time"

	"github.com/geraldleizhang/hindsight/agent/pkg/memory"
	"github.com/stretchr/testify/assert"
)

func TestDelayer(t *testing.T) {
	assert := assert.New(t)

	incoming := make(chan []memory.Trigger, 1000)

	delay := time.Duration(1) * time.Second
	delayer := delayTriggers(delay, incoming)

	select {
	case <-delayer.Outgoing:
		assert.Fail("Shouldn't be any outgoing triggers yet")
	default:
	}

	triggers := make([]memory.Trigger, 3)
	incoming <- triggers

	// Exceedingly unlikely but this could fail if >1 second elapses between the previous line of code and the next
	select {
	case <-delayer.Outgoing:
		assert.Fail("Shouldn't be any outgoing triggers yet")
	default:
	}

	time.Sleep(time.Duration(200) * time.Millisecond)

	// Exceedingly unlikely but this could fail if >0.8 second elapses between the previous line of code and the next
	select {
	case <-delayer.Outgoing:
		assert.Fail("Shouldn't be any outgoing triggers yet")
	default:
	}

	time.Sleep(time.Duration(500) * time.Millisecond)

	// Exceedingly unlikely but this could fail if >0.3 second elapses between the previous line of code and the next
	select {
	case <-delayer.Outgoing:
		assert.Fail("Shouldn't be any outgoing triggers yet")
	default:
	}

	time.Sleep(time.Duration(1) * time.Second)

	// Exceedingly unlikely but this could fail if the delayer was super slow in firing
	select {
	case <-delayer.Outgoing:
	default:
		assert.Fail("Expect to see outgoing triggers by now")
	}
}

func TestDelayerBulk(t *testing.T) {
	assert := assert.New(t)

	incoming := make(chan []memory.Trigger, 1000)

	delay := time.Duration(1) * time.Second
	delayer := delayTriggers(delay, incoming)

	select {
	case <-delayer.Outgoing:
		assert.Fail("Shouldn't be any outgoing triggers yet")
	default:
	}

	for i := 0; i < 10; i++ {
		triggers := make([]memory.Trigger, i)
		incoming <- triggers
	}

	// Exceedingly unlikely but this could fail if >1 second elapses between the previous line of code and the next
	select {
	case <-delayer.Outgoing:
		assert.Fail("Shouldn't be any outgoing triggers yet")
	default:
	}

	time.Sleep(time.Duration(500) * time.Millisecond)

	// Exceedingly unlikely but this could fail if >0.3 second elapses between the previous line of code and the next
	select {
	case <-delayer.Outgoing:
		assert.Fail("Shouldn't be any outgoing triggers yet")
	default:
	}

	time.Sleep(time.Duration(1) * time.Second)

	for i := 0; i < 10; i++ {
		// Exceedingly unlikely but this could fail if the delayer was super slow in firing
		select {
		case triggers := <-delayer.Outgoing:
			assert.Equal(i, len(triggers), "Did not received the expected outgoing triggers")
		default:
			assert.Fail("Expect to see outgoing triggers by now")
		}
	}
}

func TestDelayerStaggered(t *testing.T) {
	assert := assert.New(t)

	incoming := make(chan []memory.Trigger, 1000)

	delay := time.Duration(1) * time.Second
	delayer := delayTriggers(delay, incoming)

	select {
	case <-delayer.Outgoing:
		assert.Fail("Shouldn't be any outgoing triggers yet")
	default:
	}

	for i := 0; i < 10; i++ {
		triggers := make([]memory.Trigger, i)
		incoming <- triggers
	}

	// Exceedingly unlikely but this could fail if >1 second elapses between the previous line of code and the next
	select {
	case <-delayer.Outgoing:
		assert.Fail("Shouldn't be any outgoing triggers yet")
	default:
	}

	time.Sleep(time.Duration(500) * time.Millisecond)

	// Exceedingly unlikely but this could fail if >0.3 second elapses between the previous line of code and the next
	select {
	case <-delayer.Outgoing:
		assert.Fail("Shouldn't be any outgoing triggers yet")
	default:
	}

	for i := 10; i < 20; i++ {
		triggers := make([]memory.Trigger, i)
		incoming <- triggers
	}

	time.Sleep(time.Duration(700) * time.Millisecond)

	for i := 0; i < 10; i++ {
		// Exceedingly unlikely but this could fail if the delayer was super slow in firing
		select {
		case triggers := <-delayer.Outgoing:
			assert.Equal(i, len(triggers), "Did not received the expected outgoing triggers")
		default:
			assert.Fail("Expect to see outgoing triggers by now")
		}
	}

	// Exceedingly unlikely but this could fail if >0.3 second elapses between the previous line of code and the next
	select {
	case <-delayer.Outgoing:
		assert.Fail("Shouldn't be any outgoing triggers yet")
	default:
	}

	time.Sleep(time.Duration(500) * time.Millisecond)

	for i := 10; i < 20; i++ {
		// Exceedingly unlikely but this could fail if the delayer was super slow in firing
		select {
		case triggers := <-delayer.Outgoing:
			assert.Equal(i, len(triggers), "Did not received the expected outgoing triggers")
		default:
			assert.Fail("Expect to see outgoing triggers by now")
		}
	}
	select {
	case <-delayer.Outgoing:
		assert.Fail("Shouldn't be any outgoing triggers left")
	default:
	}
}

func TestDelayerNoTimeout(t *testing.T) {
	assert := assert.New(t)

	incoming := make(chan []memory.Trigger, 1000)

	delay := time.Duration(0) * time.Second
	delayer := delayTriggers(delay, incoming)

	for i := 0; i < 10; i++ {
		triggers := make([]memory.Trigger, i)
		incoming <- triggers
	}

	// Should be enough to allow the delayer to trigger
	time.Sleep(100 * time.Millisecond)

	for i := 0; i < 10; i++ {
		select {
		case triggers := <-delayer.Outgoing:
			assert.Equal(i, len(triggers), "Did not received the expected outgoing triggers")
		default:
			assert.Fail("Expect to see outgoing triggers by now")
		}
	}
}
