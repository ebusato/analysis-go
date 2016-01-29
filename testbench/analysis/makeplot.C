
TProfile* getProfile(TNtuple *nt, string pmt, int event, int color, int markerstyle)
{
  nt->Draw(Form("%s : time >> p%s(999,0,999)",pmt.c_str(),pmt.c_str()),Form("event==%d", event),"profgoff");
  TProfile* p = (TProfile*) gDirectory->Get(Form("p%s",pmt.c_str()));
  p->SetLineColor(color);
  p->SetMarkerStyle(markerstyle);
  p->SetMarkerColor(color);
  p->GetYaxis()->SetRangeUser(0,5000);
  return p;
}

TH2F* draw2D(TNtuple *nt, string pmt1, string pmt2, int event)
{
  nt->Draw(Form("%s : %s>>h2%s%s",pmt1.c_str(),pmt2.c_str(),pmt1.c_str(),pmt2.c_str()),Form("event==%d", event));
  TH2F* h = (TH2F*) gDirectory->Get(Form("h2%s%s",pmt1.c_str(),pmt2.c_str()));
  return h;
}

TH1F* getHisto(TNtuple *nt, string var, string cut, string hName, string xBinsRange="")
{
  nt->Draw(Form("%s>>%s%s",var.c_str(),hName.c_str(),xBinsRange.c_str()), 
	   Form("%s",cut.c_str()));
  TH1F* h = (TH1F*) gDirectory->Get(hName.c_str());
  h->GetXaxis()->SetTitle(var.c_str());
  h->SetLineWidth(2);
 return h;
}

void makePulsePlots(int event) 
{
  TNtuple* nt = new TNtuple("nt","nt","event:time:pmt1:pmt2:pmt3:pmt4");
  nt->ReadFile("output/pulses.txt");
   
  TProfile* pmt1 = getProfile(nt,"pmt1",event,kRed,24);
  TProfile* pmt2 = getProfile(nt,"pmt2",event,kGreen+2,25);
  TProfile* pmt3 = getProfile(nt,"pmt3",event,kBlue+3,26);
  TProfile* pmt4 = getProfile(nt,"pmt4",event,kOrange+1,5);
  
  TCanvas* c1 = new TCanvas("c1","c1",1000,500);
  pmt1->Draw("");
  pmt2->Draw("same");
  pmt3->Draw("same");
  pmt4->Draw("same");
  
  TCanvas* c2 = new TCanvas("c2","c2",1000,500);
  TH2F* hcorr12 = draw2D(nt,"pmt1","pmt2",event);
  double corr12 = hcorr12->GetCorrelationFactor();
  cout << "corr12=" << corr12 << endl;
}

void makeGlobalPlots()
{
  gStyle->SetOptTitle(0);
  gStyle->SetOptStat(0);
	
  TNtuple* nt = new TNtuple("nt","nt","event:pmt:hasSig:ampl:charge:srout");
  nt->ReadFile("output/globalEventVariables.txt");
  
  TCanvas* c1 = new TCanvas("c1","c1",1000,500);
  gPad->SetLogy();
  TH1F* hChargePMT0 = getHisto(nt,"charge","pmt == 0","hCharge0","(100,-100e3,600e3)");
  TH1F* hChargePMT1 = getHisto(nt,"charge","pmt == 1","hCharge1","(100,-100e3,600e3)");
  TH1F* hChargePMT2 = getHisto(nt,"charge","pmt == 2","hCharge2","(100,-100e3,600e3)");
  TH1F* hChargePMT3 = getHisto(nt,"charge","pmt == 3","hCharge3","(100,-100e3,600e3)");
  hChargePMT0->SetLineColor(kRed);
  hChargePMT1->SetLineColor(kGreen+2);
  hChargePMT2->SetLineColor(kBlue+3);
  hChargePMT3->SetLineColor(kOrange+1);
  
  hChargePMT0->Draw();
  hChargePMT1->Draw("same");
  hChargePMT2->Draw("same");
  hChargePMT3->Draw("same");
  
  TCanvas* c2 = new TCanvas("c2","c2",1000,500);
  gPad->SetLogy();
  TH1F* hAmplPMT0 = getHisto(nt,"ampl","pmt == 0","hAmpl0");
  TH1F* hAmplPMT1 = getHisto(nt,"ampl","pmt == 1","hAmpl1");
  TH1F* hAmplPMT2 = getHisto(nt,"ampl","pmt == 2","hAmpl2");
  TH1F* hAmplPMT3 = getHisto(nt,"ampl","pmt == 3","hAmpl3");
  hAmplPMT0->SetLineColor(kRed);
  hAmplPMT1->SetLineColor(kGreen+2);
  hAmplPMT2->SetLineColor(kBlue+3);
  hAmplPMT3->SetLineColor(kOrange+1);
  
  hAmplPMT0->Draw();
  hAmplPMT1->Draw("same");
  hAmplPMT2->Draw("same");
  hAmplPMT3->Draw("same");
  
  TCanvas* c3 = new TCanvas("c3","c3",1000,500);
  TH1F* hSRoutPMT0 = getHisto(nt,"srout","pmt == 0","hSRout0","(1024,0,1023)");
  TH1F* hSRoutPMT1 = getHisto(nt,"srout","pmt == 1","hSRout1","(1024,0,1023)");
  TH1F* hSRoutPMT2 = getHisto(nt,"srout","pmt == 2","hSRout2","(1024,0,1023)");
  TH1F* hSRoutPMT3 = getHisto(nt,"srout","pmt == 3","hSRout3","(1024,0,1023)");
  hSRoutPMT0->SetLineColor(kRed);
  hSRoutPMT1->SetLineColor(kGreen+2);
  hSRoutPMT2->SetLineColor(kBlue+3);
  hSRoutPMT3->SetLineColor(kOrange+1);
  
  hSRoutPMT0->Draw();
  hSRoutPMT1->Draw("same");
  hSRoutPMT2->Draw("same");
  hSRoutPMT3->Draw("same");
  
  TCanvas* c4 = new TCanvas("c4","c4",1000,500);
  TH1F* hHasSigPMT0 = getHisto(nt,"hasSig","pmt == 0","hHasSig0","(2,0,2)");
  TH1F* hHasSigPMT1 = getHisto(nt,"hasSig","pmt == 1","hHasSig1","(2,0,2)");
  TH1F* hHasSigPMT2 = getHisto(nt,"hasSig","pmt == 2","hHasSig2","(2,0,2)");
  TH1F* hHasSigPMT3 = getHisto(nt,"hasSig","pmt == 3","hHasSig3","(2,0,2)");
  hHasSigPMT0->SetLineColor(kRed);
  hHasSigPMT1->SetLineColor(kGreen+2);
  hHasSigPMT2->SetLineColor(kBlue+3);
  hHasSigPMT3->SetLineColor(kOrange+1);
  
  hHasSigPMT0->Draw();
  hHasSigPMT1->Draw("same");
  hHasSigPMT2->Draw("same");
  hHasSigPMT3->Draw("same");
}