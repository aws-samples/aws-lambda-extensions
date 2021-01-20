// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

using System;
using System.Reflection;
using System.Threading.Tasks;

namespace csharp_example_extension
{
    class Program
    {  
        static async Task Main(string[] args)
        {
            var extensionName = (1 == args.Length)
                ? args[0]
                : Assembly.GetEntryAssembly()?.GetName()?.Name;
            
            if (string.IsNullOrWhiteSpace(extensionName)) {
                throw new InvalidOperationException("Failed to determine extension name!");
            }

            using var client = new ExtensionClient(extensionName);

            // ProcessEvents will loop internally until SHUTDOWN event is received
            await client.ProcessEvents(
                // this expression will be called immediately after successful extension registration with Lambda Extension API
                onInit: async id => {
                    Console.WriteLine($"[{extensionName}] Registered extension with id = {id}");
                    await Task.CompletedTask; // useless await, so that compiler doesn't report warnings
                },
                // this will be called every time Lambda is invoked
                onInvoke: async payload =>
                {
                    Console.WriteLine($"[{extensionName}] Handling invoke from extension: {payload}");
                    await Task.CompletedTask; // useless await, so that compiler doesn't report warnings
                },
                // this will be called just once - after receiving SHUTDOWN event and before exitting the loop
                onShutdown: payload => // this is an example of lambda expression implementation without async keyword
                {
                    Console.WriteLine($"[{extensionName}] Shutting down extension: {payload}");
                    return Task.CompletedTask;
                });
        }
    }
}
