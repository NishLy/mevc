interface ISchedule {
  id: number
  room_id: number
  start: Date
  end: Date
  pattern: {
    frequency: "daily" | "weekly" | "monthly"
    interval: number
    weekday?: number[]
    monthday?: number[]
  }
}

export default ISchedule
