package lcm1602_lcd

import (
	"fmt"
	"time"

	"github.com/labstack/gommon/log"
	"golang.org/x/exp/io/i2c"
)

const (
	// commands
	LCD_CLEARDISPLAY   = 0x01
	LCD_RETURNHOME     = 0x02
	LCD_ENTRYMODESET   = 0x04
	LCD_DISPLAYCONTROL = 0x08
	LCD_CURSORSHIFT    = 0x10
	LCD_FUNCTIONSET    = 0x20
	LCD_SETCGRAMADDR   = 0x40
	LCD_SETDDRAMADDR   = 0x80

	// flags for display entry mode
	LCD_ENTRYRIGHT          = 0x00
	LCD_ENTRYLEFT           = 0x02
	LCD_ENTRYSHIFTINCREMENT = 0x01
	LCD_ENTRYSHIFTDECREMENT = 0x00

	// flags for display on/off control
	LCD_DISPLAYON  = 0x04
	LCD_DISPLAYOFF = 0x00
	LCD_CURSORON   = 0x02
	LCD_CURSOROFF  = 0x00
	LCD_BLINKON    = 0x01
	LCD_BLINKOFF   = 0x00

	// flags for display/cursor shift
	LCD_DISPLAYMOVE = 0x08
	LCD_CURSORMOVE  = 0x00
	LCD_MOVERIGHT   = 0x04
	LCD_MOVELEFT    = 0x00

	// flags for function set
	LCD_8BITMODE = 0x10
	LCD_4BITMODE = 0x00
	LCD_2LINE    = 0x08
	LCD_1LINE    = 0x00
	LCD_5x10DOTS = 0x04
	LCD_5x8DOTS  = 0x00

	// flags for backlight control
	LCD_BACKLIGHT   = 0x08
	LCD_NOBACKLIGHT = 0x00

	En = 0b00000100 // Enable bit
	Rw = 0b00000010 // Read/Write bit
	Rs = 0b00000001 // Register select bit
)

// LCM1602LCD encapsulates communication with an LCD i2c panel
type LCM1602LCD struct {
	i2c *i2c.Device
}

// NewLCM1602LCD instantiates a new LCM1602LCD for the address provided
func NewLCM1602LCD(i2c *i2c.Device) (*LCM1602LCD, error) {
	l := &LCM1602LCD{i2c}

	if err := l.initialSetup(); err != nil {
		return nil, fmt.Errorf("could not setup LCD device: %w", err)
	}

	return l, nil
}

// Low level function to write data to the i2c bus
func (l *LCM1602LCD) writeCmd(cmd byte) error {
	if err := l.i2c.Write([]byte{cmd}); err != nil {
		return err
	}
	time.Sleep(100 * time.Microsecond)
	return nil
}

// Perform initial setup of the LCD
func (l *LCM1602LCD) initialSetup() error {
	initBytes := []byte{0x03, 0x03, 0x03, 0x02}
	for _, b := range initBytes {
		if err := l.lcdWrite(b, 0); err != nil {
			return err
		}
	}

	setupBytes := []byte{
		LCD_FUNCTIONSET | LCD_2LINE | LCD_5x8DOTS | LCD_4BITMODE,
		LCD_DISPLAYCONTROL | LCD_DISPLAYON,
		LCD_CLEARDISPLAY,
		LCD_ENTRYMODESET | LCD_ENTRYLEFT,
	}
	for _, b := range setupBytes {
		if err := l.lcdWrite(b, 0); err != nil {
			return err
		}
	}

	time.Sleep(200 * time.Millisecond)

	return nil
}

func (l *LCM1602LCD) lcdWrite4bits(cmd byte) error {
	if err := l.writeCmd(cmd | LCD_BACKLIGHT); err != nil {
		return err
	}
	return l.lcdStrobe(cmd)
}

func (l *LCM1602LCD) lcdWrite(cmd byte, mode byte) error {
	if err := l.lcdWrite4bits(mode | (cmd & 0xF0)); err != nil {
		return err
	}
	return l.lcdWrite4bits(mode | ((cmd << 4) & 0xF0))
}

// Clear clears the LCD display
func (l *LCM1602LCD) Clear() error {
	if err := l.lcdWrite(LCD_CLEARDISPLAY, En); err != nil {
		return err
	}

	return l.lcdWrite(LCD_RETURNHOME, En)
}

// clocks EN to latch command
func (l *LCM1602LCD) lcdStrobe(data byte) error {
	if err := l.writeCmd(data | En | LCD_BACKLIGHT); err != nil {
		return err
	}

	time.Sleep(500 * time.Microsecond)

	if err := l.writeCmd((data &^ En) | LCD_BACKLIGHT); err != nil {
		return err
	}

	time.Sleep(100 * time.Microsecond)

	return nil
}

// WriteString writes a string to the LCD at the given row
func (l *LCM1602LCD) WriteString(message string, row int, startPosition byte) error {
	log.Debugf("Will write string to LCD: %s", message)
	var position byte

	switch row {
	case 1:
		position = startPosition
	case 2:
		position = 0x40 + startPosition
	case 3:
		position = 0x14 + startPosition
	case 4:
		position = 0x54 + startPosition
	}

	l.lcdWrite(0x80+position, 0)

	for _, c := range []byte(message) {
		if err := l.lcdWrite(c, Rs); err != nil {
			return err
		}
	}

	return nil
}
