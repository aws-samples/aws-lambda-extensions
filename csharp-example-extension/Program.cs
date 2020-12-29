// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

using System;
using System.Threading.Tasks;

namespace example_extension
{
    class Program
    {  
        static async Task Main(string[] args)
        {
            using var client = new ExtensionClient();

            // ProcessEvents will loop internally until SHUTDOWN event is received
            await client.ProcessEvents(
                // this expression will be called immediately after successful extension registration with Lambda Extension API
                onInit: async id => {
                    Console.WriteLine($"Registered extension with id = {id}");
                    await Task.CompletedTask; // useless await, so that compiler doesn't report warnings
                },
                // this will be called every time Lambda is invoked
                onInvoke: async payload =>
                {
                    Console.WriteLine($"Handling invoke from extension: {payload}");
                    await Task.CompletedTask; // useless await, so that compiler doesn't report warnings
                },
                // this will be called just once - after receiving SHUTDOWN event and before exitting the loop
                onShutdown: payload => // this is an example of lambda expression implementation without async keyword
                {
                    Console.WriteLine($"Shutting down extension: {payload}");
                    return Task.CompletedTask;
                });
        }
    }
}
