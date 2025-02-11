use image::{self, GenericImageView};
use std::{env, fs::File, io::{self, BufRead, Write}, path::PathBuf};

fn png_to_huh(png_path: PathBuf, huh_path: PathBuf) -> Result<(), std::io::Error> {
    let img = image::open(&png_path).expect("PNG file not found!");

    let width = img.width();
    let height = img.height();

    let mut output = vec![];

    output.extend_from_slice(&width.to_ne_bytes());
    output.extend_from_slice(&height.to_ne_bytes());

    for pixel in img.pixels() {
        output.push(pixel.2[0]); 
        output.push(pixel.2[1]); 
        output.push(pixel.2[2]); 
    }

    let mut file = File::create(huh_path)?;
    file.write_all(&output)?;
    file.flush()?;

    println!("PNG file has been successfully converted to HUH format!");
    Ok(())
}

fn main() {
    let args: Vec<String> = env::args().collect();

    if args.len() == 3 {
        let png_path: PathBuf = args[1].clone().into();
        let huh_path: PathBuf = args[2].clone().into();

        if let Err(e) = png_to_huh(png_path, huh_path) {
            eprintln!("Conversion error: {:?}", e);
        }
    } else {
        println!("Console arguments are missing. Moving from the file.txt file to reading...");

        let file_path = "file.txt";

        let file = File::open(file_path).expect("file.txt is not found!");
        let mut lines = io::BufReader::new(file).lines();

        let input_file = lines.next().expect("PNG file name is missing ").expect("Reading error");
        let output_file = lines.next().expect("HUH file name is missing").expect("Reading error");

        let png_path: PathBuf = input_file.into();
        let huh_path: PathBuf = output_file.into();

        if let Err(e) = png_to_huh(png_path, huh_path) {
            eprintln!("Conversion error: {:?}", e);
        }
    }
}