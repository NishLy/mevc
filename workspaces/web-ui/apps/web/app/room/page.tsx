import Room from "@/features/stream/components/room"

export default async function Page({
  searchParams,
}: {
  searchParams: Promise<{ [key: string]: string | string[] | undefined }>
}) {
  // Await the searchParams object
  const filters = await searchParams
  const roomId = filters.roomId

  return (
    <div>
      {roomId ? (
        <Room roomId={typeof roomId === "string" ? roomId : "default-room"} />
      ) : (
        <p>No room ID provided</p>
      )}
    </div>
  )
}
