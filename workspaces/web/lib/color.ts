class ColorGenerator {
  private hue: number
  private goldenRatioConjugate: number
  private saturation: number
  private lightness: number

  constructor(saturation = 70, lightness = 60) {
    this.hue = Math.random() // Start at a random point
    this.goldenRatioConjugate = 0.618033988749895
    this.saturation = saturation
    this.lightness = lightness
  }

  /**
   * Generates the next distinct HEX color in the sequence.
   */
  next() {
    this.hue += this.goldenRatioConjugate
    this.hue %= 1 // Keep it between 0 and 1
    return this._hslToHex(this.hue * 360, this.saturation, this.lightness)
  }

  /**
   * Helper to convert HSL values to a HEX string.
   */
  private _hslToHex(h: number, s: number, l: number): string {
    l /= 100
    const a = (s * Math.min(l, 1 - l)) / 100
    const f = (n: number) => {
      const k = (n + h / 30) % 12
      const color = l - a * Math.max(Math.min(k - 3, 9 - k, 1), -1)
      return Math.round(255 * color)
        .toString(16)
        .padStart(2, "0")
    }
    return `#${f(0)}${f(8)}${f(4)}`.toUpperCase()
  }

  /**
   * Generates an array of N distinct colors.
   */
  generateBatch(count: number) {
    return Array.from({ length: count }, () => this.next())
  }
}

// --- Usage ---
const ColorGen = new ColorGenerator()

export default ColorGen
