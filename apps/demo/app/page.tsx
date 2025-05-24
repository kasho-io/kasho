import Image from "next/image";

export default function Home() {
  return (
    <div className="min-h-screen p-8 bg-base-200">
      <div className="max-w-4xl mx-auto space-y-8">        
        {/* Card with badge */}
        <div className="card bg-base-100 shadow-xl text-base-content">
          <div className="card-body">
            <div className="flex items-center gap-2">
              <h2 className="card-title">Database Status</h2>
              <div className="badge badge-success">Connected</div>
            </div>
            <p>Primary and replica databases are synchronized.</p>
          </div>
        </div>

        {/* Buttons with different variants */}
        <div className="flex gap-4 justify-center">
          <button className="btn btn-primary">Primary</button>
          <button className="btn btn-secondary">Secondary</button>
          <button className="btn btn-accent">Accent</button>
        </div>

        {/* Alert */}
        <div className="alert alert-info text-base-content">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" className="stroke-current shrink-0 w-6 h-6">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
          </svg>
          <span>Changes in the primary database will be reflected in real-time in the replica.</span>
        </div>

        {/* Table with JetBrains Mono font */}
        <div className="overflow-x-auto">
          <table className="table table-zebra w-full font-mono">
            <thead>
              <tr>
                <th>ID</th>
                <th>Name</th>
                <th>Status</th>
                <th>Value</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>1</td>
                <td>Alice</td>
                <td>Active</td>
                <td>42.7</td>
              </tr>
              <tr>
                <td>2</td>
                <td>Bob</td>
                <td>Inactive</td>
                <td>13.2</td>
              </tr>
              <tr>
                <td>3</td>
                <td>Charlie</td>
                <td>Active</td>
                <td>99.9</td>
              </tr>
              <tr>
                <td>4</td>
                <td>Dana</td>
                <td>Pending</td>
                <td>7.5</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
