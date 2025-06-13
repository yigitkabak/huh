use image::{self};
use std::{env, fs::File, io::{self, Write, Read}, path::PathBuf};
use std::error::Error;
use std::time::Duration;
use crossterm::{
    execute,
    style::{Color, Print, ResetColor, SetForegroundColor},
    terminal::{enable_raw_mode, disable_raw_mode},
    event::{self, Event, KeyCode, KeyEvent},
};
use viuer::{Config, print_from_file};

const LOGO: &str = r#"
  _   _ _   _ _   _   _____                           _            
 | | | | | | | | | | /  __ \                         | |           
 | |_| | | | | |_| | | /  \/ ___  _ ____   _____ _ __| |_ ___ _ __ 
 |  _  | | | |  _  | | |    / _ \| '_ \ \ / / _ \ '__| __/ _ \ '__|
 | | | | |_| | | | | | \__/\ (_) | | | \ V /  __/ |  | ||  __/ |   
 \_| |_/\___/\_| |_|  \____/\___/|_| |_|\_/ \___|_|   \__\___|_|   
"#;

fn print_fancy_header() {
    execute!(
        io::stdout(),
        SetForegroundColor(Color::Cyan),
        Print(LOGO),
        Print("\n"),
        ResetColor
    ).unwrap();
    
    println!("   🖼️  Universal Image Converter & Viewer  🖼️\n");
}

fn print_progress(progress: f32) {
    let bar_width = 50;
    let filled = (progress * bar_width as f32) as usize;
    let empty = bar_width - filled;
    
    let bar: String = "█".repeat(filled) + &" ".repeat(empty);
    
    execute!(
        io::stdout(),
        SetForegroundColor(Color::Green),
        Print(format!("\r[{}] {:.1}%", bar, progress * 100.0)),
        ResetColor
    ).unwrap();
    io::stdout().flush().unwrap();
}

fn print_success(message: &str) {
    execute!(
        io::stdout(),
        SetForegroundColor(Color::Green),
        Print(format!("\n✅ {}\n", message)),
        ResetColor
    ).unwrap();
}

fn print_error(message: &str) {
    execute!(
        io::stdout(),
        SetForegroundColor(Color::Red),
        Print(format!("\n❌ {}\n", message)),
        ResetColor
    ).unwrap();
}

fn print_info(message: &str) {
    execute!(
        io::stdout(),
        SetForegroundColor(Color::Blue),
        Print(format!("\nℹ️ {}\n", message)),
        ResetColor
    ).unwrap();
}

fn image_to_huh(image_path: &PathBuf, huh_path: &PathBuf) -> Result<(), Box<dyn Error>> {
    print_info(&format!("Converting {} to {}", image_path.display(), huh_path.display()));
    
    let img = image::open(image_path)?.to_rgb8();
    let width = img.width();
    let height = img.height();
    let total_pixels = (width * height) as usize;
    
    let mut output = Vec::new();
    
    // Write header
    output.extend_from_slice(&width.to_ne_bytes());
    output.extend_from_slice(&height.to_ne_bytes());
    
    for (i, pixel) in img.pixels().enumerate() {
        output.push(pixel[0]);
        output.push(pixel[1]);
        output.push(pixel[2]);
        
        if i % (total_pixels / 100 + 1) == 0 {
            print_progress(i as f32 / total_pixels as f32);
        }
    }
    
    print_progress(1.0);
    println!();
    
    let mut file = File::create(huh_path)?;
    file.write_all(&output)?;
    file.flush()?;
    
    print_success(&format!("Successfully converted {} to {}", image_path.display(), huh_path.display()));
    Ok(())
}

fn huh_to_image(huh_path: &PathBuf, image_path: &PathBuf) -> Result<(), Box<dyn Error>> {
    print_info(&format!("Converting {} to {}", huh_path.display(), image_path.display()));
    
    let mut file = File::open(huh_path)?;
    let mut buffer = Vec::new();
    file.read_to_end(&mut buffer)?;
    
    if buffer.len() < 8 {
        return Err("Invalid HUH file: header too small".into());
    }
    
    let width = u32::from_ne_bytes([buffer[0], buffer[1], buffer[2], buffer[3]]);
    let height = u32::from_ne_bytes([buffer[4], buffer[5], buffer[6], buffer[7]]);
    
    let total_pixels = (width * height) as usize;
    if buffer.len() != 8 + total_pixels * 3 {
        return Err(format!("Invalid HUH file: expected {} bytes, got {}", 8 + total_pixels * 3, buffer.len()).into());
    }
    
    let mut img = image::RgbImage::new(width, height);
    
    for y in 0..height {
        for x in 0..width {
            let pos = 8 + 3 * (y * width + x) as usize;
            let r = buffer[pos];
            let g = buffer[pos + 1];
            let b = buffer[pos + 2];
            img.put_pixel(x, y, image::Rgb([r, g, b]));
            
            if (y * width + x) % (total_pixels as u32 / 100 + 1) == 0 {
                print_progress((y * width + x) as f32 / total_pixels as f32);
            }
        }
    }
    
    print_progress(1.0);
    println!();
    
    img.save(image_path)?;
    
    print_success(&format!("Successfully converted {} to {}", huh_path.display(), image_path.display()));
    Ok(())
}

fn view_image(path: &PathBuf) -> Result<(), Box<dyn Error>> {
    print_info(&format!("Viewing image: {}", path.display()));
    
    let ext = path.extension().unwrap_or_default().to_string_lossy().to_string();
    
    if ext == "huh" {
        let temp_path = PathBuf::from("temp_view.png");
        huh_to_image(path, &temp_path)?;
        
        let config = Config {
            premultiplied_alpha: false, // Eklendi
            width: None,
            height: None,
            absolute_offset: false,
            x: 0,
            y: 0,
            restore_cursor: true,
            use_kitty: true,
            use_iterm: true,
            transparent: false,
            truecolor: true,
        };
        
        print_from_file(&temp_path, &config)?;
        
        std::fs::remove_file(temp_path)?;
    } else {
        let config = Config {
            premultiplied_alpha: false, // Eklendi
            width: None,
            height: None,
            absolute_offset: false,
            x: 0,
            y: 0,
            restore_cursor: true,
            use_kitty: true,
            use_iterm: true,
            transparent: false,
            truecolor: true,
        };
        
        print_from_file(path, &config)?;
    }
    
    println!("\nPress 'q' to exit viewer...");
    enable_raw_mode()?;
    
    loop {
        if event::poll(Duration::from_millis(100))? {
            if let Event::Key(KeyEvent { code, .. }) = event::read()? {
                if code == KeyCode::Char('q') || code == KeyCode::Esc {
                    break;
                }
            }
        }
    }
    
    disable_raw_mode()?;
    Ok(())
}

fn print_usage() {
    print_fancy_header();
    println!("Usage:");
    println!("  huh convert <input_file> <output_file>  - Convert between image formats and HUH");
    println!("  huh view <file>                        - View an image or HUH file");
    println!("  huh help                               - Show this help message");
    println!("\nExamples:");
    println!("  huh convert image.png image.huh         - Convert PNG to HUH");
    println!("  huh convert image.huh image.jpg         - Convert HUH to JPG");
    println!("  huh view image.png                      - View a PNG image");
    println!("  huh view image.huh                      - View a HUH file");
}

fn main() -> Result<(), Box<dyn Error>> {
    let args: Vec<String> = env::args().collect();
    
    if args.len() < 2 {
        print_usage();
        return Ok(());
    }
    
    match args[1].as_str() {
        "convert" => {
            if args.len() != 4 {
                print_error("Invalid number of arguments for convert command");
                print_usage();
                return Ok(());
            }
            
            let input_path: PathBuf = args[2].clone().into();
            let output_path: PathBuf = args[3].clone().into();
            
            let input_ext = input_path.extension().unwrap_or_default().to_string_lossy().to_string();
            let output_ext = output_path.extension().unwrap_or_default().to_string_lossy().to_string();
            
            if !input_path.exists() {
                print_error(&format!("Input file does not exist: {}", input_path.display()));
                return Ok(());
            }
            
            if input_ext == "huh" && output_ext != "huh" {
                huh_to_image(&input_path, &output_path)?;
            } else if input_ext != "huh" && output_ext == "huh" {
                image_to_huh(&input_path, &output_path)?;
            } else if input_ext != "huh" && output_ext != "huh" {
                print_info(&format!("Converting {} to {}", input_path.display(), output_path.display()));
                let img = image::open(&input_path)?;
                img.save(&output_path)?;
                print_success(&format!("Successfully converted {} to {}", input_path.display(), output_path.display()));
            } else {
                print_error("Cannot convert from HUH to HUH");
            }
        },
        "view" => {
            if args.len() != 3 {
                print_error("Invalid number of arguments for view command");
                print_usage();
                return Ok(());
            }
            
            let file_path: PathBuf = args[2].clone().into();
            
            if !file_path.exists() {
                print_error(&format!("File does not exist: {}", file_path.display()));
                return Ok(());
            }
            
            view_image(&file_path)?;
        },
        "help" | "--help" | "-h" => {
            print_usage();
        },
        _ => {
            print_error(&format!("Unknown command: {}", args[1]));
            print_usage();
        }
    }
    
    Ok(())
}
