import React, { useState } from "react";

export default function AuthPage({
    setAuthenticated,
}: {
    setAuthenticated: (authenticated: boolean) => void;
}) {
    const [isLoginView, setIsLoginView] = useState(true);
    const [formData, setFormData] = useState({
        username: "",
        email: "",
        password: "",
    });
    const [error, setError] = useState("");
    const [loading, setLoading] = useState(false);

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setFormData({ ...formData, [e.target.name]: e.target.value });
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError("");
        setLoading(true);

        const endpoint = isLoginView ? "/api/auth/login" : "/api/auth/register";

        try {
            const response = await fetch(`http://localhost:3000${endpoint}`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(
                    isLoginView
                        ? { email: formData.email, password: formData.password }
                        : formData,
                ),
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.error || "Something went wrong");
            }

            localStorage.setItem("jwt_token", data.token);
            setAuthenticated(true);

            alert(`${isLoginView ? "Login" : "Registration"} successful!`);
        } catch (err: any) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div
            style={{
                maxWidth: "400px",
                margin: "50px auto",
                fontFamily: "sans-serif",
            }}
        >
            <h2>{isLoginView ? "Log In" : "Create an Account"}</h2>

            {error && (
                <div style={{ color: "red", marginBottom: "10px" }}>
                    {error}
                </div>
            )}

            <form
                onSubmit={handleSubmit}
                style={{
                    display: "flex",
                    flexDirection: "column",
                    gap: "15px",
                }}
            >
                {!isLoginView && (
                    <div>
                        <label>Username</label>
                        <br />
                        <input
                            type="text"
                            name="username"
                            value={formData.username}
                            onChange={handleChange}
                            required
                            style={{ width: "100%", padding: "8px" }}
                        />
                    </div>
                )}

                <div>
                    <label>Email</label>
                    <br />
                    <input
                        type="email"
                        name="email"
                        value={formData.email}
                        onChange={handleChange}
                        required
                        style={{ width: "100%", padding: "8px" }}
                    />
                </div>

                <div>
                    <label>Password</label>
                    <br />
                    <input
                        type="password"
                        name="password"
                        value={formData.password}
                        onChange={handleChange}
                        required
                        style={{ width: "100%", padding: "8px" }}
                    />
                </div>

                <button
                    type="submit"
                    disabled={loading}
                    style={{ padding: "10px", cursor: "pointer" }}
                >
                    {loading
                        ? "Processing..."
                        : isLoginView
                          ? "Log In"
                          : "Register"}
                </button>
            </form>

            <div style={{ marginTop: "20px", textAlign: "center" }}>
                <button
                    onClick={() => {
                        setIsLoginView(!isLoginView);
                        setError("");
                    }}
                    style={{
                        background: "none",
                        border: "none",
                        color: "blue",
                        cursor: "pointer",
                        textDecoration: "underline",
                    }}
                >
                    {isLoginView
                        ? "Need an account? Register here"
                        : "Already have an account? Log in"}
                </button>
            </div>
        </div>
    );
}
