export interface MenuItemDef {
  icon: React.ReactNode
  label: string
  variant?: "danger"
  onClick?: () => void
}

export interface Participant {
  id: string
  name: string
  initials: string
  color?: string
  role: "Host" | "Guest"
}
