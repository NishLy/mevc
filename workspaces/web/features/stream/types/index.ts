import ISchedule from "@/features/schedule/types"
import IUser from "@/types/user"

interface IRoom {
  id: number
  code: string // Unique 8-character code
  name: string
  description?: string
  hostId: string
  url: string
  host: IUser
  pin?: string // Exclude hashed pin from JSON responses
  schedules?: ISchedule[]
  auto_join: boolean
  allow_guests: boolean
  allow_recording: boolean
  allow_chat: boolean
  allow_screen_share: boolean
  capacity: number
  location?: string
}

export default IRoom
