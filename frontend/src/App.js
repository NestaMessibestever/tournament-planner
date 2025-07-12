// Main App component - This is the root component of our React application
// For now, it's a simple placeholder that proves React is working

import React from 'react';

function App() {
  return (
    <div className="min-h-screen bg-gray-50">
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-4xl font-bold text-center text-indigo-600 mb-8">
          Tournament Planner
        </h1>
        
        <div className="max-w-2xl mx-auto bg-white rounded-lg shadow-md p-6">
          <h2 className="text-2xl font-semibold mb-4">Welcome!</h2>
          <p className="text-gray-600 mb-4">
            Your Tournament Planner is now running successfully! This is a placeholder page
            that confirms React is working properly.
          </p>
          
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-4">
            <h3 className="font-semibold text-blue-900 mb-2">What's Working:</h3>
            <ul className="list-disc list-inside text-blue-800 space-y-1">
              <li>React is running on port 3000</li>
              <li>Tailwind CSS is configured for styling</li>
              <li>The basic project structure is in place</li>
            </ul>
          </div>
          
          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
            <h3 className="font-semibold text-yellow-900 mb-2">Next Steps:</h3>
            <ul className="list-disc list-inside text-yellow-800 space-y-1">
              <li>Add the backend Go server code</li>
              <li>Implement authentication components</li>
              <li>Create the tournament management features</li>
              <li>Connect frontend to backend API</li>
            </ul>
          </div>
          
          <div className="mt-6 text-center">
            <p className="text-sm text-gray-500">
              Backend API should be running on{' '}
              <a href="http://localhost:8080" className="text-indigo-600 hover:underline">
                http://localhost:8080
              </a>
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}

export default App;