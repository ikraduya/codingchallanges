"use server";

type State = { short_url: string; error?: string };

export async function createShortURL(
  prevState: State,
  formData: FormData
): Promise<State> {
  const longURL = formData.get("url");

  if (typeof longURL !== "string" || !longURL) {
    return { short_url: "", error: "URL is required" };
  }

  try {
    const res = await fetch(process.env.BACKEND_URL ?? "", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ url: longURL }),
      cache: "no-store",
    });

    if (!res.ok) {
      const text = await res.text().catch(() => "");
      return { short_url: "", error: text || "Failed to shorten URL" };
    }

    const data = await res.json();
    return { short_url: data?.short_url ?? "" };
  } catch (error) {
    return { short_url: "", error: "Server error" };
  }
}
