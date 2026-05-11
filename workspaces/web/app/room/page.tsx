import Preload from "@/features/stream/components/preload"
import { ErrorBoundary } from "next/dist/client/components/error-boundary"
import { Suspense } from "react"

export default async function Page({
  searchParams,
}: {
  searchParams: Promise<{ [key: string]: string | string[] | undefined }>
}) {
  // Await the searchParams object
  const filters = await searchParams

  const code = filters.code

  return (
    <div>
      {code ? (
        <Suspense fallback={<div>Loading room details...</div>}>
          <Preload code={code as string} />
        </Suspense>
      ) : (
        <p>No code provided</p>
      )}
    </div>
  )
}
