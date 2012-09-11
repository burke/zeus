require 'sinatra'

get '/?' do
  redirect '/here' 
end

get '/here/?' do
  erb :here
end

get '/easy/?' do
  erb :easy
end

get '/frequently/questionable/?' do
  erb :faq
end

get '/open/source/?' do
  erb :contributing
end

get '/configurable/?' do
  erb :configuring
end

get '/documented?' do
  erb :documentation
end
