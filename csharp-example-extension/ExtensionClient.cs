// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

using System;
using System.Reflection;
using System.Net.Http;
using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;
using System.Linq;
using System.Threading.Tasks;
using System.Threading;

namespace example_extension
{
    /// <summary>
    /// Lambda Extension API client
    /// </summary>
    internal class ExtensionClient : IDisposable
    {
        #region HTTP header key names

        /// <summary>
        /// HTTP header that is used to register a new extension name with Extension API
        /// </summary>
        private const string LambdaExtensionNameHeader = "Lambda-Extension-Name";

        /// <summary>
        /// HTTP header used to provide extension registration id
        /// </summary>
        /// <remarks>
        /// Registration endpoint reply will have this header value with a new id, assigned to this extension by the API.
        /// All other endpoints will expect HTTP calls to have id header attached to all requests.
        /// </remarks>
        private const string LambdaExtensionIdHeader = "Lambda-Extension-Identifier";

        /// <summary>
        /// HTTP header to report Lambda Extension error type string.
        /// </summary>
        /// <remarks>
        /// This header is used to report additional error details for Init and Shutdown errors.
        /// </remarks>
        private const string LambdaExtensionFunctionErrorTypeHeader = "Lambda-Extension-Function-Error-Type";

        #endregion

        #region Environment variable names

        /// <summary>
        /// Environment variable that holds server name and port number for Extension API endpoints
        /// </summary>
        private const string LambdaRuntimeApiAddress = "AWS_LAMBDA_RUNTIME_API";

        #endregion

        #region Instance properties

        /// <summary>
        /// Extension id, which is assigned to this extension after the registration
        /// </summary>
        public string Id { get; private set; }

        #endregion

        #region Constructor and readonly variables

        /// <summary>
        /// Http client instance
        /// </summary>
        /// <remarks>This is an IDisposable object that must be properly disposed of,
        /// thus <see cref="ExtensionClient"/> implements <see cref="IDisposable"/> interface too.</remarks>
        private readonly HttpClient httpClient = new HttpClient();

        /// <summary>
        /// Extension name, calculated from the current executing assembly name
        /// </summary>
        private readonly string extensionName = Assembly.GetExecutingAssembly().GetName().Name;

        /// <summary>
        /// Extension registration URL
        /// </summary>
        private readonly Uri registerUrl;

        /// <summary>
        /// Next event long poll URL
        /// </summary>
        private readonly Uri nextUrl;

        /// <summary>
        /// Extension initialization error reporting URL
        /// </summary>
        private readonly Uri initErrorUrl;

        /// <summary>
        /// Extension shutdown error reporting URL
        /// </summary>
        private readonly Uri shutdownErrorUrl;

        /// <summary>
        /// Constructor
        /// </summary>
        public ExtensionClient()
        {
            // Set infinite timeout so that underlying connection is kept alive
            this.httpClient.Timeout = Timeout.InfiniteTimeSpan;
            // Get Extension API service base URL from the environment variable
            var apiUri = new UriBuilder(Environment.GetEnvironmentVariable(LambdaRuntimeApiAddress)).Uri;
            // Common path for all Extension API URLs
            var basePath = "2020-01-01/extension";

            // Calculate all Extension API endpoints' URLs
            this.registerUrl = new Uri(apiUri, $"{basePath}/register");
            this.nextUrl = new Uri(apiUri, $"{basePath}/event/next");
            this.initErrorUrl = new Uri(apiUri, $"{basePath}/init/error");
            this.shutdownErrorUrl = new Uri(apiUri, $"{basePath}/exit/error");
        }

        #endregion

        #region Public interface

        /// <summary>
        /// Extension registration and event loop handling
        /// </summary>
        /// <param name="onInit">Optional lambda extension that is invoked when extension has been successfully registered with AWS Lambda Extension API.
        /// This function will be called exactly once if it is defined and ignored if this parameter is null.</param>
        /// <param name="onInvoke">Optional lambda extension that is invoked every time AWS Lambda Extension API reports a new <see cref="ExtensionEvent.INVOKE"/> event.
        /// This function will be called once for each <see cref="ExtensionEvent.INVOKE"/> event during the entire lifetime of AWS Lambda function instance.</param>
        /// <param name="onShutdown">Optional lambda extension that is invoked when extension receives <see cref="ExtensionEvent.SHUTDOWN"/> event from AWS LAmbda Extension API.
        /// This function will be called exactly once if it is defined and ignored if this parameter is null.</param>
        /// <returns>Awaitable void</returns>
        /// <remarks>Unhandled exceptions thrown by <paramref name="onInit"/> and <paramref name="onShutdown"/> functions will be reported to AWS Lambda API with
        /// <c>/init/error</c> and <c>/exit/error</c> calls, in any case <see cref="ProcessEvents"/> will immediately exit after reporting the error.
        /// Unhandled <paramref name="onInvoke"/> exceptions are logged to console and ignored, so that extension execution can continue.
        /// </remarks>
        public async Task ProcessEvents(Func<string, Task> onInit = null, Func<string, Task> onInvoke = null, Func<string, Task> onShutdown = null)
        {
            // Register extension with AWS Lambda Extension API to handle both INVOKE and SHUTDOWN events
            await RegisterExtensionAsync(ExtensionEvent.INVOKE, ExtensionEvent.SHUTDOWN);

            // If onInit function is defined, invoke it and report any unhandled exceptions
            if (!await SafeInvoke(onInit, this.Id, ex => ReportErrorAsync(this.initErrorUrl, "Fatal.Unhandled", ex))) return;

            // loop till SHUTDOWN event is received
            var hasNext = true;
            while (hasNext)
            {
                // get the next event type and details
                var (type, payload) = await GetNextAsync();

                switch (type)
                {
                    case ExtensionEvent.INVOKE:
                        // invoke onInit function if one is defined and log unhandled exceptions
                        // event loop will continue even if there was an exception
                        await SafeInvoke(onInvoke, payload, onException: ex => {
                            Console.WriteLine("Invoke handler threw an exception");
                            return Task.CompletedTask;
                        });
                        break;
                    case ExtensionEvent.SHUTDOWN:
                        // terminate the loop, invoke onShutdown function if there is any and report any unhandled exceptions to AWS Extension API
                        hasNext = false;
                        await SafeInvoke(onShutdown, this.Id, ex => ReportErrorAsync(this.shutdownErrorUrl, "Fatal.Unhandled", ex));
                        break;
                    default:
                        throw new ApplicationException($"Unexpected event type: {type}");
                }
            }
        }

        #endregion

        #region Private methods

        /// <summary>
        /// Register extension with Extension API
        /// </summary>
        /// <param name="events">Event types to by notified with</param>
        /// <returns>Awaitable void</returns>
        /// <remarks>This method is expected to be called just once when extension is being registered with the Extension API.</remarks>
        private async Task RegisterExtensionAsync(params ExtensionEvent[] events)
        {
            // custom options for JsonSerializer to serialize ExtensionEvent enum values as strings, rather than integers
            // thus we produce strongly typed code, which doesn't rely on strings
            var options = new JsonSerializerOptions();
            options.Converters.Add(new JsonStringEnumConverter());

            // create Json content for this extension registration
            using var content = new StringContent(JsonSerializer.Serialize(new {
                events
            }, options), Encoding.UTF8, "application/json");

            // add extension name header value
            content.Headers.Add(LambdaExtensionNameHeader, this.extensionName);

            // POST call to Extension API
            using var response = await this.httpClient.PostAsync(this.registerUrl, content);

            // if POST call didn't succeed
            if (!response.IsSuccessStatusCode)
            {
                // log details
                Console.WriteLine($"Error response received for registration request: {await response.Content.ReadAsStringAsync()}");
                // throw an unhandled exception, so that extension is terminated by Lambda runtime
                response.EnsureSuccessStatusCode();
            }

            // get registration id from the response header
            this.Id = response.Headers.GetValues(LambdaExtensionIdHeader).FirstOrDefault();
            // if registration id is empty
            if (string.IsNullOrEmpty(this.Id))
            {
                // throw an exception
                throw new ApplicationException("Extension API register call didn't return a valid identifier.");
            }
            // configure all HttpClient to send registration id header along with all subsequent requests
            this.httpClient.DefaultRequestHeaders.Add(LambdaExtensionIdHeader, this.Id);
        }

        /// <summary>
        /// Long poll for the next event from Extension API
        /// </summary>
        /// <returns>Awaitable tuple having event type and event details fields</returns>
        /// <remarks>It is important to have httpClient.Timeout set to some value, that is longer than any expected wait time,
        /// otherwise HttpClient will throw an exception when getting the next event details from the server.</remarks>
        private async Task<(ExtensionEvent type, string payload)> GetNextAsync()
        {
            // use GET request to long poll for the next event
            var contentBody = await this.httpClient.GetStringAsync(this.nextUrl);

            // use JsonDocument instead of JsonSerializer, since there is no need to construct the entire object
            using var doc = JsonDocument.Parse(contentBody);

            // extract eventType from the reply, convert it to ExtensionEvent enum and reply with the typed event type and event content details.
            return new (Enum.Parse<ExtensionEvent>(doc.RootElement.GetProperty("eventType").GetString()), contentBody);
        }

        /// <summary>
        /// Report initialization or shutdown error
        /// </summary>
        /// <param name="url"><see cref="initErrorUrl"/> or <see cref="shutdownErrorUrl"/></param>
        /// <param name="errorType">Error type string, e.g. Fatal.ConnectionError or any other meaningful type</param>
        /// <param name="exception">Exception details</param>
        /// <returns>Awaitable void</returns>
        /// <remarks>This implementation will append <paramref name="exception"/> name to <paramref name="errorType"/> for demonstration purposes</remarks>
        private async Task ReportErrorAsync(Uri url, string errorType, Exception exception)
        {
            using var content = new StringContent(string.Empty);
            content.Headers.Add(LambdaExtensionIdHeader, this.Id);
            content.Headers.Add(LambdaExtensionFunctionErrorTypeHeader, $"{errorType}.{exception.GetType().Name}");

            using var response = await this.httpClient.PostAsync(url, content);
            if (!response.IsSuccessStatusCode)
            {
                Console.WriteLine($"Error response received for {url.PathAndQuery}: {await response.Content.ReadAsStringAsync()}");
                response.EnsureSuccessStatusCode();
            }
        }

        /// <summary>
        /// Try to invoke <paramref name="func"/> and call <paramref name="onException"/> if <paramref name="func"/> threw an exception
        /// </summary>
        /// <param name="func">Function to be invoked. Do nothing if it is null.</param>
        /// <param name="param">Parameter to pass to the <paramref name="func"/></param>
        /// <param name="onException">Exception reporting function to be called in case of an exception. Can be null.</param>
        /// <returns>Awaitable boolean value. True if <paramref name="func"/> succeeded and False otherwise.</returns>
        private async Task<bool> SafeInvoke(Func<string, Task> func, string param, Func<Exception, Task> onException)
        {
            try
            {
                await func?.Invoke(param);
                return true;
            }
            catch (Exception ex)
            {
                await onException?.Invoke(ex);
                return false;
            }
        }

        #endregion

        #region IDisposable implementation

        /// <summary>
        /// Dispose of instance Disposable variables
        /// </summary>
        public void Dispose()
        {
            // Quick and dirty implementation to propagate Dispose call to HttpClient instance
            ((IDisposable)httpClient).Dispose();
        }

        #endregion
    }
}