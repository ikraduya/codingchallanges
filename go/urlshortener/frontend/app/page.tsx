"use client";

import { useActionState, useState } from "react";
import { useFormState, useFormStatus } from "react-dom";
import { createShortURL } from "./actions";

const initialState = { short_url: "", error: "" };

export default function Home() {
  const { pending } = useFormStatus();
  const [state, formAction] = useActionState(createShortURL, initialState);

  const shortLinkClasses = `bg-white/5 border border-dashed border-white/20 rounded-xl transition-all duration-300 ease-out overflow-hidden ${
    state.short_url
      ? "opacity-100 translate-y-0 p-6 mb-8 visible max-h-64"
      : "opacity-0 -translate-y-2 invisible max-h-0 pointer-events-none"
  }`;

  const copyToClipboard = async () => {
    if (state.short_url) await navigator.clipboard.writeText(state.short_url);
  };

  return (
    <div className="bg-gradient-to-br from-slate-900 to-blue-900 min-h-screen flex items-center justify-center p-6">
      <div className="max-w-2xl w-full bg-white/10 backdrop-blur-lg rounded-2xl shadow-2xl px-8 pt-8 border border-white/20">
        <div className="mb-8 text-center">
          <h1 className="text-3xl font-bold text-white mb-2">
            Coding Challenges URL Shortener
          </h1>
          <p className="text-blue-200">
            Paste your long URL below to create a short, shareable link.
          </p>
        </div>

        <form
          className="flex flex-col md:flex-row gap-3 mb-8"
          action={formAction}
        >
          <input
            name="url"
            type="url"
            placeholder="https://example.com/very-long-link-address"
            className="flex-1 px-5 py-4 rounded-xl bg-white/5 border border-white/10 text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 transition-all"
            required
          />
          <button
            disabled={pending}
            type="submit"
            className="bg-blue-600 hover:bg-blue-500 text-white font-semibold px-8 py-4 rounded-xl transition-all active:scale-95 shadow-lg shadow-blue-900/20 cursor-pointer"
          >
            {pending ? "Shortening..." : "Shorten"}
          </button>
        </form>

        <div className={shortLinkClasses}>
          <p className="text-xs uppercase tracking-widest text-blue-300 font-bold mb-3">
            Your Short Link
          </p>
          <div className="flex items-center justify-between bg-black/20 p-4 rounded-lg group">
            <span
              id="short-url"
              className="text-blue-400 font-mono text-lg break-all"
            >
              {state.short_url}
            </span>
            <button
              onClick={copyToClipboard}
              className="ml-4 text-slate-400 hover:text-white transition-colors p-2"
              title="Copy to clipboard"
              type="button"
            >
              <i className="fa-regular fa-copy text-xl cursor-pointer">Copy</i>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
