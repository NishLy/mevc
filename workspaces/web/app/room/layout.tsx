import QueryProviders from "@/providers/query"

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return <QueryProviders>{children}</QueryProviders>
}
