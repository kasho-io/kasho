"use client";

import React, { Component, ReactNode } from "react";

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export default class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): State {
    // Update state so the next render will show the fallback UI
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    // Log error to console in development
    if (process.env.NODE_ENV === "development") {
      console.error("ErrorBoundary caught an error:", error, errorInfo);
    }
    // In production, you might want to log to an error reporting service
  }

  render() {
    if (this.state.hasError) {
      // Custom fallback UI
      if (this.props.fallback) {
        return this.props.fallback;
      }

      // Default error UI
      return (
        <div className="alert alert-error" role="alert">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-6 w-6 shrink-0 stroke-current"
            fill="none"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth="2"
              d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
          <div>
            <h3 className="font-bold">Something went wrong</h3>
            <div className="text-sm">
              {process.env.NODE_ENV === "development" && this.state.error ? (
                <details className="mt-2">
                  <summary className="cursor-pointer">Show error details</summary>
                  <pre className="mt-2 text-xs overflow-auto">{this.state.error.message}</pre>
                </details>
              ) : (
                "Please try refreshing the page or contact support if the problem persists."
              )}
            </div>
          </div>
          <button
            className="btn btn-sm"
            onClick={() => this.setState({ hasError: false, error: null })}
            aria-label="Try again"
          >
            Try Again
          </button>
        </div>
      );
    }

    return this.props.children;
  }
}
