using System.Net.Http.Json;
using Microsoft.AspNetCore.Mvc;

namespace EdgeSight.Frontend.Controllers;

[ApiController]
[Route("api/query")]
public class QueryController : ControllerBase
{
    private readonly IHttpClientFactory _httpClientFactory;
    private readonly IConfiguration _config;

    public QueryController(IHttpClientFactory httpClientFactory, IConfiguration config)
    {
        _httpClientFactory = httpClientFactory;
        _config = config;
    }

    [HttpGet]
    public async Task<IActionResult> Get([FromQuery] string q, [FromQuery] string? location)
    {
        if (string.IsNullOrWhiteSpace(q)) return BadRequest("missing q");

        var apiBase = _config.GetValue<string>("EDGE_API_BASE") ?? "http://localhost:8080/api/v1";
        var loc = string.IsNullOrWhiteSpace(location) ? "Los Angeles" : location;

        var client = _httpClientFactory.CreateClient();
        var url = $"{apiBase}/query?q={Uri.EscapeDataString(q)}&location={Uri.EscapeDataString(loc)}";

        try
        {
            var resp = await client.GetAsync(url, HttpContext.RequestAborted);
            if (!resp.IsSuccessStatusCode)
            {
                return StatusCode((int)resp.StatusCode, await resp.Content.ReadAsStringAsync());
            }
            var payload = await resp.Content.ReadFromJsonAsync<object>(cancellationToken: HttpContext.RequestAborted);
            return Ok(payload);
        }
        catch (Exception ex)
        {
            return StatusCode(502, $"proxy error: {ex.Message}");
        }
    }
}
