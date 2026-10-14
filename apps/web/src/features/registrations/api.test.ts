import { afterEach, describe, expect, it, vi } from "vitest";
import { submitRegistration } from "./api";

describe("registration submission request", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("sends answers and IMAGE files as multipart without overriding the browser boundary", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          success: true,
          message: "Pendaftaran berhasil dikirim",
          data: { submission_id: 42, submitted_at: "2026-10-13T00:00:00Z", registration_title: "Famget" },
        }),
        { status: 201, headers: { "Content-Type": "application/json" } },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);
    const file = new File([new Uint8Array([1, 2, 3])], "bukti.png", { type: "image/png" });
    const answers = { full_name: "Ryuu", nim: "23E010005", email: "ryuu@example.test", phone: "08157840371" };

    const result = await submitRegistration(9, answers, { payment_proof: file });

    expect(result.submission_id).toBe(42);
    expect(fetchMock).toHaveBeenCalledOnce();
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toContain("/public/registrations/9/submissions");
    expect(init.method).toBe("POST");
    expect(init.credentials).toBe("include");
    expect(init.headers).not.toHaveProperty("Content-Type");
    expect(init.body).toBeInstanceOf(FormData);
    const body = init.body as FormData;
    expect(body.get("answers")).toBe(JSON.stringify(answers));
    const attached = body.get("image:payment_proof") as File;
    expect(attached.name).toBe(file.name);
    expect(attached.type).toBe(file.type);
    expect(attached.size).toBe(file.size);
    expect([...body.keys()]).toEqual(["answers", "image:payment_proof"]);
  });

  it("keeps text-only registrations as JSON", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ success: true, message: "OK", data: { submission_id: 43 } }), {
        status: 201,
        headers: { "Content-Type": "application/json" },
      }),
    );
    vi.stubGlobal("fetch", fetchMock);
    const answers = { full_name: "Ryuu", nim: "23E010006", email: "ryuu@example.test", phone: "08157840371" };

    await submitRegistration(9, answers);

    const [, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(new Headers(init.headers).get("Content-Type")).toBe("application/json");
    expect(init.body).toBe(JSON.stringify({ answers }));
  });
});
